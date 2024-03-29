package compiler

import (
	"fmt"
	"golox-lang/lib/chunk"
	"golox-lang/lib/chunk/opcode"
	"golox-lang/lib/compiler/precedence"
	"golox-lang/lib/config"
	"golox-lang/lib/debug"
	"golox-lang/lib/scanner"
	"golox-lang/lib/scanner/token"
	"golox-lang/lib/scanner/token/tokentype"
	"golox-lang/lib/value"
	"golox-lang/lib/value/valuetype"
	"os"
	"strconv"
)

type Parser struct {
	Current         token.Token
	Previous        token.Token
	HadError        bool
	PanicMode       bool
	CurrentCompiler *Compiler
	CurrentClass    *ClassCompiler

	scanner *scanner.Scanner
}

type Compiler struct {
	enclosing *Compiler
	function  *value.ObjFunction
	funcType  FunctionType

	Locals     []Local
	ScopeDepth int
}

type ClassCompiler struct {
	enclosing     *ClassCompiler
	HasSuperClass bool
}

type Local struct {
	Name  token.Token
	depth int
}

type FunctionType byte

const (
	TYPE_FUNCTION FunctionType = iota
	TYPE_INITIALIZER
	TYPE_METHOD
	TYPE_SCRIPT
)

type ParseFn func(receiver *Parser, canAssign bool)

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence precedence.Precedence
}

var rules map[tokentype.TokenType]ParseRule

func New(scanner *scanner.Scanner) *Parser {
	parser := new(Parser)
	parser.HadError = false
	parser.PanicMode = false
	parser.scanner = scanner

	return parser
}

func init() {
	rules = make(map[tokentype.TokenType]ParseRule)
	rules[tokentype.TOKEN_LEFT_PAREN] = ParseRule{(*Parser).grouping, (*Parser).call, precedence.PREC_CALL}
	rules[tokentype.TOKEN_RIGHT_PAREN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_LEFT_BRACE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_RIGHT_BRACE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_COMMA] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_DOT] = ParseRule{nil, (*Parser).dot, precedence.PREC_CALL}
	rules[tokentype.TOKEN_MINUS] = ParseRule{(*Parser).unary, (*Parser).binary, precedence.PREC_TERM}
	rules[tokentype.TOKEN_PLUS] = ParseRule{nil, (*Parser).binary, precedence.PREC_TERM}
	rules[tokentype.TOKEN_SEMICOLON] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_SLASH] = ParseRule{nil, (*Parser).binary, precedence.PREC_FACTOR}
	rules[tokentype.TOKEN_STAR] = ParseRule{nil, (*Parser).binary, precedence.PREC_FACTOR}
	rules[tokentype.TOKEN_BANG] = ParseRule{(*Parser).unary, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_BANG_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_EQUALITY}
	rules[tokentype.TOKEN_EQUAL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_EQUAL_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_EQUALITY}
	rules[tokentype.TOKEN_GREATER] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_GREATER_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_LESS] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_LESS_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_IDENTIFIER] = ParseRule{(*Parser).variable, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_STRING] = ParseRule{(*Parser).string_, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_NUMBER] = ParseRule{(*Parser).number, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_AND] = ParseRule{nil, (*Parser).and_, precedence.PREC_AND}
	rules[tokentype.TOKEN_CLASS] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_ELSE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FALSE] = ParseRule{(*Parser).literal, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FOR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FUN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_IF] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_NIL] = ParseRule{(*Parser).literal, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_OR] = ParseRule{nil, (*Parser).or, precedence.PREC_OR}
	rules[tokentype.TOKEN_PRINT] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_RETURN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_SUPER] = ParseRule{(*Parser).super, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_THIS] = ParseRule{(*Parser).this, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_TRUE] = ParseRule{(*Parser).literal, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_VAR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_WHILE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_ERROR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_EOF] = ParseRule{nil, nil, precedence.PREC_NONE}
}

func Compile(source string) *value.ObjFunction {
	scanner := scanner.New(source)

	parser := New(scanner)
	parser.initCompiler(TYPE_SCRIPT)

	parser.advance()

	for !parser.match(tokentype.TOKEN_EOF) {
		parser.declaration()
	}

	function := parser.endCompiler()

	if parser.HadError {
		return nil
	}
	return function
}

func (parser *Parser) initCompiler(funcType FunctionType) *Compiler {
	compiler := new(Compiler)
	compiler.enclosing = parser.CurrentCompiler
	compiler.function = value.NewFunction(chunk.New())
	compiler.funcType = funcType
	compiler.ScopeDepth = 0
	compiler.Locals = make([]Local, 0)

	parser.CurrentCompiler = compiler

	if funcType != TYPE_SCRIPT {
		compiler.function.Name = value.NewObjString(parser.Previous.Lexeme).AsString()
	}

	var local Local

	if funcType != TYPE_FUNCTION {
		local = Local{depth: 0, Name: token.Token{Lexeme: "this"}}
	} else {
		local = Local{depth: 0, Name: token.Token{Lexeme: ""}}
	}
	compiler.Locals = append(compiler.Locals, local)

	return compiler
}

func (parser *Parser) advance() {
	parser.Previous = parser.Current

	for {
		parser.Current = parser.scanner.ScanToken()
		if parser.Current.Type != tokentype.TOKEN_ERROR {
			break
		}

		parser.errorAtCurrent(parser.Current.Lexeme)
	}
}

func (parser *Parser) consume(tokenType tokentype.TokenType, message string) {
	if parser.Current.Type == tokenType {
		parser.advance()
		return
	}

	parser.errorAtCurrent(message)
}

func (parser *Parser) check(tokenType tokentype.TokenType) bool {
	return parser.Current.Type == tokenType
}

func (parser *Parser) match(tokenType tokentype.TokenType) bool {
	if !parser.check(tokenType) {
		return false
	}
	parser.advance()
	return true
}

func (parser *Parser) emitByte(b byte) {
	parser.currentChunk().WriteChunk(b, parser.Previous.Line)
}

func (parser *Parser) emitBytes(b1 byte, b2 byte) {
	parser.emitByte(b1)
	parser.emitByte(b2)
}

func (parser *Parser) emitLoop(loopStart int) {
	parser.emitByte(byte(opcode.OP_LOOP))

	offset := len(parser.currentChunk().GetCode()) - loopStart + 2
	// if (offset > UINT16_MAX) error("Loop body too large.");

	parser.emitByte(byte((offset >> 8) & 0xff))
	parser.emitByte(byte(offset & 0xff))
}

func (parser *Parser) emitJump(instruction opcode.OpCode) int {
	parser.emitByte(byte(instruction))
	parser.emitByte(0xff)
	parser.emitByte(0xff)
	return len(parser.currentChunk().GetCode()) - 2
}

func (parser *Parser) emitReturn() {
	if parser.CurrentCompiler.funcType == TYPE_INITIALIZER {
		parser.emitBytes(byte(opcode.OP_GET_LOCAL), 0)
	} else {
		parser.emitByte(byte(opcode.OP_NIL))
	}

	parser.emitByte(byte(opcode.OP_RETURN))
}

func (parser *Parser) emitConstant(value value.Value) {
	parser.currentChunk().WriteConstant(value, parser.Previous.Line)
}

func (parser *Parser) patchJump(offset int) {
	// -2 to adjust for the bytecode for the jump offset itself.
	jump := len(parser.currentChunk().GetCode()) - offset - 2

	// if (jump > UINT16_MAX) {
	// 	error("Too much code to jump over.");
	//   }

	parser.currentChunk().GetCode()[offset] = byte((jump >> 8) & 0xff)
	parser.currentChunk().GetCode()[offset+1] = byte(jump & 0xff)
}

func (parser *Parser) endCompiler() *value.ObjFunction {
	parser.emitReturn()
	function := parser.CurrentCompiler.function

	if config.DEBUG_PRINT_CODE {
		if !parser.HadError {
			var name string
			if function.Name != nil {
				name = function.Name.String
			} else {
				name = "<script>"
			}
			debug.DisassembleChunk(parser.currentChunk(), name)
		}
	}

	parser.CurrentCompiler = parser.CurrentCompiler.enclosing
	return function
}

func (parser *Parser) beginScope() {
	parser.CurrentCompiler.ScopeDepth++
}

func (parser *Parser) endScope() {
	parser.CurrentCompiler.ScopeDepth--

	for len(parser.CurrentCompiler.Locals) > 0 && parser.CurrentCompiler.Locals[len(parser.CurrentCompiler.Locals)-1].depth > parser.CurrentCompiler.ScopeDepth {
		parser.emitByte(byte(opcode.OP_POP))
		parser.CurrentCompiler.Locals = parser.CurrentCompiler.Locals[:len(parser.CurrentCompiler.Locals)-1]
	}
}

func (parser *Parser) binary(canAssign bool) {
	operatorType := parser.Previous.Type

	// Compile the right operand
	rule := parser.getRule(operatorType)
	parser.parsePrecedence(precedence.Precedence(rule.Precedence + 1))

	switch operatorType {
	case tokentype.TOKEN_BANG_EQUAL:
		parser.emitBytes(byte(opcode.OP_EQUAL), byte(opcode.OP_NOT))

	case tokentype.TOKEN_EQUAL_EQUAL:
		parser.emitByte(byte(opcode.OP_EQUAL))

	case tokentype.TOKEN_GREATER:
		parser.emitByte(byte(opcode.OP_GREATER))

	case tokentype.TOKEN_GREATER_EQUAL:
		parser.emitBytes(byte(opcode.OP_LESS), byte(opcode.OP_NOT))

	case tokentype.TOKEN_LESS:
		parser.emitByte(byte(opcode.OP_LESS))

	case tokentype.TOKEN_LESS_EQUAL:
		parser.emitBytes(byte(opcode.OP_GREATER), byte(opcode.OP_NOT))

	case tokentype.TOKEN_PLUS:
		parser.emitByte(byte(opcode.OP_ADD))

	case tokentype.TOKEN_MINUS:
		parser.emitBytes(byte(opcode.OP_NEGATE), byte(opcode.OP_ADD))

	case tokentype.TOKEN_STAR:
		parser.emitByte(byte(opcode.OP_MULTIPLY))

	case tokentype.TOKEN_SLASH:
		parser.emitByte(byte(opcode.OP_DIVIDE))

	default:
		return
	}
}

func (parser *Parser) call(canAssign bool) {
	argCount := parser.argumentList()
	parser.emitBytes(byte(opcode.OP_CALL), argCount)
}

func (parser *Parser) dot(canAssign bool) {
	parser.consume(tokentype.TOKEN_IDENTIFIER, "Expect property name after '.'.")
	name := parser.identifierConstant(&parser.Previous)

	if canAssign && parser.match(tokentype.TOKEN_EQUAL) {
		parser.expression()
		parser.emitLongOrShort(name, byte(opcode.OP_SET_PROPERTY), byte(opcode.OP_SET_PROPERTY_LONG))
	} else {
		parser.emitLongOrShort(name, byte(opcode.OP_GET_PROPERTY), byte(opcode.OP_GET_PROPERTY_LONG))
	}
}

func (parser *Parser) literal(canAssign bool) {
	switch parser.Previous.Type {
	case tokentype.TOKEN_FALSE:
		parser.emitByte(byte(opcode.OP_FALSE))

	case tokentype.TOKEN_NIL:
		parser.emitByte(byte(opcode.OP_NIL))

	case tokentype.TOKEN_TRUE:
		parser.emitByte(byte(opcode.OP_TRUE))

	default:
		return
	}
}

func (parser *Parser) grouping(canAssign bool) {
	parser.expression()
	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (parser *Parser) number(canAssign bool) {
	val, err := strconv.ParseFloat(parser.Previous.Lexeme, 64)
	if err != nil {
	}
	parser.emitConstant(value.New(valuetype.VAL_NUMBER, val))
}

func (parser *Parser) or(canAssign bool) {
	elseJump := parser.emitJump(opcode.OP_JUMP_IF_FALSE)
	endJump := parser.emitJump(opcode.OP_JUMP)

	parser.patchJump(elseJump)
	parser.emitByte(byte(opcode.OP_POP))

	parser.parsePrecedence(precedence.PREC_OR)
	parser.patchJump(endJump)
}

func (parser *Parser) string_(canAssign bool) {
	parser.emitConstant(
		value.NewObjString(parser.Previous.Lexeme[1 : len(parser.Previous.Lexeme)-1]))
}

func (parser *Parser) namedVariable(name token.Token, canAssign bool) {
	var getOp, setOp opcode.OpCode
	var getOpLong, setOpLong opcode.OpCode
	arg := parser.resolveLocal(&name)

	if arg != -1 {
		getOp = opcode.OP_GET_LOCAL
		setOp = opcode.OP_SET_LOCAL
		getOpLong = opcode.OP_GET_LOCAL_LONG
		setOpLong = opcode.OP_SET_LOCAL_LONG
	} else {
		arg = parser.identifierConstant(&name)
		getOp = opcode.OP_GET_GLOBAL
		setOp = opcode.OP_SET_GLOBAL
		getOpLong = opcode.OP_GET_GLOBAL_LONG
		setOpLong = opcode.OP_SET_GLOBAL_LONG
	}

	if canAssign && parser.match(tokentype.TOKEN_EQUAL) {
		parser.expression()
		parser.emitLongOrShort(arg, byte(setOp), byte(setOpLong))
	} else {
		parser.emitLongOrShort(arg, byte(getOp), byte(getOpLong))
	}
}

func (parser *Parser) emitLongOrShort(val int, shortOp byte, longOp byte) {
	if val < 256 {
		parser.emitBytes(shortOp, byte(val))
	} else {
		parser.emitByte(longOp)
		parser.emitByte(byte(val & 0xff))
		parser.emitByte(byte((val >> 8) & 0xff))
		parser.emitByte(byte((val >> 16) & 0xff))
	}
}

func (parser *Parser) variable(canAssign bool) {
	parser.namedVariable(parser.Previous, canAssign)
}

func (parser *Parser) syntheticToken(text string) token.Token {
	return token.Token{Lexeme: text}
}

func (parser *Parser) super(canAssign bool) {
	if parser.CurrentClass == nil {
		parser.error("Can't use 'super' outside of a class.")
	} else if !parser.CurrentClass.HasSuperClass {
		parser.error("Can't use 'super' in a class with no superclass.")
	}

	parser.consume(tokentype.TOKEN_DOT, "Expect '.' after 'super'.")
	parser.consume(tokentype.TOKEN_IDENTIFIER, "Expect superclass method name.")
	name := parser.identifierConstant(&parser.Previous)

	parser.namedVariable(parser.syntheticToken("this"), false)
	parser.namedVariable(parser.syntheticToken("super"), false)
	parser.emitBytes(byte(opcode.OP_GET_SUPER), byte(name))
}

func (parser *Parser) this(canAssign bool) {
	if parser.CurrentClass == nil {
		parser.error("Can't use 'this' outside of a class.")
		return
	}

	parser.variable(false)
}

func (parser *Parser) unary(canAssign bool) {
	operatorType := parser.Previous.Type

	// Compile the operand
	parser.parsePrecedence(precedence.PREC_UNARY)

	switch operatorType {
	case tokentype.TOKEN_BANG:
		parser.emitByte(byte(opcode.OP_NOT))

	case tokentype.TOKEN_MINUS:
		parser.emitByte(byte(opcode.OP_NEGATE))

	default:
		return
	}
}

func (parser *Parser) parsePrecedence(preced precedence.Precedence) {
	parser.advance()
	prefixRule := parser.getRule(parser.Previous.Type).Prefix
	if prefixRule == nil {
		parser.error("Expect Expression.")
		return
	}

	canAssign := preced <= precedence.PREC_ASSIGNMENT
	prefixRule(parser, canAssign)

	for preced <= parser.getRule(parser.Current.Type).Precedence {
		parser.advance()
		infixRule := parser.getRule(parser.Previous.Type).Infix
		infixRule(parser, canAssign)
	}

	if canAssign && parser.match(tokentype.TOKEN_EQUAL) {
		parser.error("Invalid assignmet target.")
	}
}

func (parser *Parser) identifierConstant(name *token.Token) int {
	return parser.currentChunk().AddConstant(value.NewObjString(name.Lexeme))
}

func (parser *Parser) resolveLocal(name *token.Token) int {
	for i := len(parser.CurrentCompiler.Locals) - 1; i >= 0; i-- {
		local := &parser.CurrentCompiler.Locals[i]
		if name.Lexeme == local.Name.Lexeme {
			if local.depth == -1 {
				parser.error("Cannot read local variable in its own initializer.")
			}
			return i
		}
	}

	return -1
}

func (parser *Parser) addLocal(name token.Token) {
	local := Local{Name: name, depth: -1}
	parser.CurrentCompiler.Locals = append(parser.CurrentCompiler.Locals, local)
}

func (parser *Parser) declareVariable() {
	// Global variables are implicitly declared.
	if parser.CurrentCompiler.ScopeDepth == 0 {
		return
	}

	name := &parser.Previous
	for i := len(parser.CurrentCompiler.Locals) - 1; i >= 0; i-- {
		local := &parser.CurrentCompiler.Locals[i]
		if local.depth != -1 && local.depth < parser.CurrentCompiler.ScopeDepth {
			break
		}

		if name.Lexeme == local.Name.Lexeme {
			parser.error("Variable with this name already declared in this scope.")
		}
	}
	parser.addLocal(*name)
}

func (parser *Parser) parserVariable(errorMessage string) int {
	parser.consume(tokentype.TOKEN_IDENTIFIER, errorMessage)

	parser.declareVariable()
	if parser.CurrentCompiler.ScopeDepth > 0 {
		return 0
	}

	return parser.identifierConstant(&parser.Previous)
}

func (parser *Parser) markInitialized() {
	if parser.CurrentCompiler.ScopeDepth == 0 {
		return
	}
	parser.CurrentCompiler.Locals[len(parser.CurrentCompiler.Locals)-1].depth = parser.CurrentCompiler.ScopeDepth
}

func (parser *Parser) defineVariable(global int) {
	if parser.CurrentCompiler.ScopeDepth > 0 {
		parser.markInitialized()
		return
	}

	parser.emitLongOrShort(global, byte(opcode.OP_DEFINE_GLOBAL), byte(opcode.OP_DEFINE_GLOBAL_LONG))
}

func (parser *Parser) argumentList() byte {
	var argCount byte = 0
	if !parser.check(tokentype.TOKEN_RIGHT_PAREN) {
		for {
			parser.expression()

			if argCount == 255 {
				parser.error("Cant have more than 255 arguments.")
			}
			argCount++
			if !parser.match(tokentype.TOKEN_COMMA) {
				break
			}
		}
	}

	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after arguments.")
	return argCount
}

func (parser *Parser) and_(canAssign bool) {
	endJump := parser.emitJump(opcode.OP_JUMP_IF_FALSE)

	parser.emitByte(byte(opcode.OP_POP))
	parser.parsePrecedence(precedence.PREC_AND)

	parser.patchJump(endJump)
}

func (parser *Parser) getRule(token tokentype.TokenType) *ParseRule {
	rule := rules[token]
	return &rule
}

func (parser *Parser) expression() {
	parser.parsePrecedence(precedence.PREC_ASSIGNMENT)
}

func (parser *Parser) block() {
	for !parser.check(tokentype.TOKEN_RIGHT_BRACE) && !parser.check(tokentype.TOKEN_EOF) {
		parser.declaration()
	}

	parser.consume(tokentype.TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (parser *Parser) function(funcType FunctionType) {
	parser.initCompiler(funcType)
	parser.beginScope()

	// Compile the parameter list.
	parser.consume(tokentype.TOKEN_LEFT_PAREN, "Expect '(' after function name.")
	if !parser.check(tokentype.TOKEN_RIGHT_PAREN) {
		for {
			parser.CurrentCompiler.function.Arity++
			if parser.CurrentCompiler.function.Arity > 255 {
				parser.errorAtCurrent("Cant have more than 255 parameters.")
			}

			paramConstant := parser.parserVariable("Expect parameter name.")
			parser.defineVariable(paramConstant)

			if !parser.match(tokentype.TOKEN_COMMA) {
				break
			}
		}
	}
	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after parameters.")

	// The body
	parser.consume(tokentype.TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	parser.block()

	// Create the function object.
	function := parser.endCompiler()
	parser.emitConstant(value.NewObjFunction(function))
}

func (parser *Parser) method() {
	parser.consume(tokentype.TOKEN_IDENTIFIER, "Expect method name.")
	constant := parser.identifierConstant(&parser.Previous)

	funcType := TYPE_METHOD
	if parser.Previous.Lexeme == "init" {
		funcType = TYPE_INITIALIZER
	}
	parser.function(funcType)

	parser.emitLongOrShort(constant, byte(opcode.OP_METHOD), byte(opcode.OP_METHOD_LONG))
}

func (parser *Parser) classDeclaration() {
	parser.consume(tokentype.TOKEN_IDENTIFIER, "Expect class name.")
	className := parser.Previous
	nameConstant := parser.identifierConstant(&parser.Previous)
	parser.declareVariable()

	parser.emitLongOrShort(nameConstant, byte(opcode.OP_CLASS), byte(opcode.OP_CLASS_LONG))

	parser.defineVariable(nameConstant)

	var classCompiler ClassCompiler
	classCompiler.HasSuperClass = false
	classCompiler.enclosing = parser.CurrentClass
	parser.CurrentClass = &classCompiler

	if parser.match(tokentype.TOKEN_LESS) {
		parser.consume(tokentype.TOKEN_IDENTIFIER, "Expect superclass name.")
		parser.variable(false)

		if className.Lexeme == parser.Previous.Lexeme {
			parser.error("A class can't inherit from itself.")
		}

		parser.beginScope()
		parser.addLocal(parser.syntheticToken("super"))
		parser.defineVariable(0)

		parser.namedVariable(className, false)
		parser.emitByte(byte(opcode.OP_INHERIT))
		classCompiler.HasSuperClass = true
	}

	parser.namedVariable(className, false)
	parser.consume(tokentype.TOKEN_LEFT_BRACE, "Expect '{' before class body.")

	for !parser.check(tokentype.TOKEN_RIGHT_BRACE) && !parser.check(tokentype.TOKEN_EOF) {
		parser.method()
	}

	parser.consume(tokentype.TOKEN_RIGHT_BRACE, "Expect '}' after class body.")
	parser.emitByte(byte(opcode.OP_POP))

	if classCompiler.HasSuperClass {
		parser.endScope()
	}

	parser.CurrentClass = parser.CurrentClass.enclosing
}

func (parser *Parser) funDeclaration() {
	global := parser.parserVariable("Expect function name.")
	parser.markInitialized()
	parser.function(TYPE_FUNCTION)
	parser.defineVariable(global)
}

func (parser *Parser) varDeclaration() {
	global := parser.parserVariable("Expect variable name.")

	if parser.match(tokentype.TOKEN_EQUAL) {
		parser.expression()
	} else {
		parser.emitByte(byte(opcode.OP_NIL))
	}
	parser.consume(tokentype.TOKEN_SEMICOLON, "Expect ';' after variable declaration.")

	parser.defineVariable(global)
}

func (parser *Parser) expressionStatement() {
	parser.expression()
	parser.consume(tokentype.TOKEN_SEMICOLON, "Expect ';' after expression.")
	parser.emitByte(byte(opcode.OP_POP))
}

func (parser *Parser) forStatement() {
	// Variables declared in the initializer should be scoped to the loop body
	parser.beginScope()

	parser.consume(tokentype.TOKEN_LEFT_PAREN, "Expect '(' after 'for'.")
	if parser.match(tokentype.TOKEN_SEMICOLON) {
		// No initializer
	} else if parser.match(tokentype.TOKEN_VAR) {
		parser.varDeclaration()
	} else {
		parser.expressionStatement()
	}

	loopStart := len(parser.currentChunk().GetCode())

	exitJump := -1
	if !parser.match(tokentype.TOKEN_SEMICOLON) {
		parser.expression()
		parser.consume(tokentype.TOKEN_SEMICOLON, "Expect ';' after loop condition.")

		// Jump out of the loop if the condition is false.
		exitJump = parser.emitJump(opcode.OP_JUMP_IF_FALSE)
		parser.emitByte(byte(opcode.OP_POP))
	}

	if !parser.match(tokentype.TOKEN_RIGHT_PAREN) {
		bodyJump := parser.emitJump(opcode.OP_JUMP)

		incrementStart := len(parser.currentChunk().GetCode())
		parser.expression()
		parser.emitByte(byte(opcode.OP_POP))
		parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")

		parser.emitLoop(loopStart)
		loopStart = incrementStart
		parser.patchJump(bodyJump)
	}

	parser.statement()

	parser.emitLoop(loopStart)

	if exitJump != -1 {
		parser.patchJump(exitJump)
		parser.emitByte(byte(opcode.OP_POP))
	}

	parser.endScope()
}

func (parser *Parser) ifStatement() {
	parser.consume(tokentype.TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	parser.expression()
	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after condition")

	thenJump := parser.emitJump(opcode.OP_JUMP_IF_FALSE)
	parser.emitByte(byte(opcode.OP_POP))
	parser.statement()

	elseJump := parser.emitJump(opcode.OP_JUMP)

	parser.patchJump(thenJump)
	parser.emitByte(byte(opcode.OP_POP))

	if parser.match(tokentype.TOKEN_ELSE) {
		parser.statement()
	}
	parser.patchJump(elseJump)
}

func (parser *Parser) printStatement() {
	parser.expression()
	parser.consume(tokentype.TOKEN_SEMICOLON, "Expect ';' after value.")
	parser.emitByte(byte(opcode.OP_PRINT))
}

func (parser *Parser) returnStatement() {
	if parser.CurrentCompiler.funcType == TYPE_SCRIPT {
		parser.error("Cant return from top-level code.")
	}

	if parser.match(tokentype.TOKEN_SEMICOLON) {
		parser.emitReturn()
	} else {
		if parser.CurrentCompiler.funcType == TYPE_INITIALIZER {
			parser.error("Can't return a value from an initializer.")
		}

		parser.expression()
		parser.consume(tokentype.TOKEN_SEMICOLON, "Expect ';' after return value.")
		parser.emitByte(byte(opcode.OP_RETURN))
	}
}

func (parser *Parser) whileStatement() {
	loopStart := len(parser.currentChunk().GetCode())

	parser.consume(tokentype.TOKEN_LEFT_PAREN, "Expect '(' after 'while'.")
	parser.expression()
	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := parser.emitJump(opcode.OP_JUMP_IF_FALSE)

	parser.emitByte(byte(opcode.OP_POP))
	parser.statement()

	parser.emitLoop(loopStart)

	parser.patchJump(exitJump)
	parser.emitByte(byte(opcode.OP_POP))
}

func (parser *Parser) synchronize() {
	parser.PanicMode = false
	for parser.Current.Type != tokentype.TOKEN_EOF {
		if parser.Previous.Type == tokentype.TOKEN_SEMICOLON {
			return
		}

		switch parser.Current.Type {
		case tokentype.TOKEN_CLASS:
		case tokentype.TOKEN_FUN:
		case tokentype.TOKEN_VAR:
		case tokentype.TOKEN_FOR:
		case tokentype.TOKEN_IF:
		case tokentype.TOKEN_WHILE:
		case tokentype.TOKEN_PRINT:
		case tokentype.TOKEN_RETURN:
			return

		default:
			// Do nothing.

		}

		parser.advance()
	}
}

func (parser *Parser) declaration() {
	if parser.match(tokentype.TOKEN_CLASS) {
		parser.classDeclaration()
	} else if parser.match(tokentype.TOKEN_FUN) {
		parser.funDeclaration()
	} else if parser.match(tokentype.TOKEN_VAR) {
		parser.varDeclaration()
	} else {
		parser.statement()
	}

	if parser.PanicMode {
		parser.synchronize()
	}
}

func (parser *Parser) statement() {
	if parser.match(tokentype.TOKEN_PRINT) {
		parser.printStatement()
	} else if parser.match(tokentype.TOKEN_FOR) {
		parser.forStatement()
	} else if parser.match(tokentype.TOKEN_IF) {
		parser.ifStatement()
	} else if parser.match(tokentype.TOKEN_RETURN) {
		parser.returnStatement()
	} else if parser.match(tokentype.TOKEN_WHILE) {
		parser.whileStatement()
	} else if parser.match(tokentype.TOKEN_LEFT_BRACE) {
		parser.beginScope()
		parser.block()
		parser.endScope()
	} else {
		parser.expressionStatement()
	}
}

func (parser *Parser) currentChunk() *chunk.Chunk {
	return parser.CurrentCompiler.function.Chunk.(*chunk.Chunk)
}

func (parser *Parser) errorAt(token *token.Token, message string) {
	if parser.PanicMode {
		return
	}
	parser.PanicMode = true

	fmt.Fprintf(os.Stderr, "[line %d] Error", token.Line)

	if token.Type == tokentype.TOKEN_EOF {
		fmt.Fprintf(os.Stderr, " at end")
	} else if token.Type == tokentype.TOKEN_ERROR {

	} else {
		fmt.Fprintf(os.Stderr, " at %s", token.Lexeme)
	}

	fmt.Fprintf(os.Stderr, ": %s\n", message)
	parser.HadError = true
}

func (parser *Parser) error(message string) {
	parser.errorAt(&parser.Previous, message)
}

func (parser *Parser) errorAtCurrent(message string) {
	parser.errorAt(&parser.Current, message)
}
