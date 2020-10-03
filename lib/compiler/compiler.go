package compiler

import (
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/compiler/precedence"
	"golanglox/lib/config"
	"golanglox/lib/debug"
	"golanglox/lib/scanner"
	"golanglox/lib/scanner/token"
	"golanglox/lib/scanner/token/tokentype"
	"golanglox/lib/value"
	"golanglox/lib/value/valuetype"
	"os"
	"strconv"
)

// Parser struct
type Parser struct {
	Current   token.Token
	Previous  token.Token
	HadError  bool
	PanicMode bool
	CurrentC  *Compiler

	scanner *scanner.Scanner
	chunk   *chunk.Chunk
}

// Compiler struct
type Compiler struct {
	Locals     []Local
	ScopeDepth int
}

// Local struct
type Local struct {
	Name  token.Token
	depth int
}

// ParseFn func
type ParseFn func(receiver *Parser, canAssign bool)

// ParseRule struct
type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence precedence.Precedence
}

var rules map[tokentype.TokenType]ParseRule

// New creates a new parser and returns it
func New(scanner *scanner.Scanner, chunk *chunk.Chunk, compiler *Compiler) *Parser {
	parser := new(Parser)
	parser.HadError = false
	parser.PanicMode = false
	parser.scanner = scanner
	parser.chunk = chunk
	parser.CurrentC = compiler
	return parser
}

func newCompiler() *Compiler {
	compiler := new(Compiler)
	compiler.ScopeDepth = 0
	compiler.Locals = make([]Local, 0)
	return compiler
}

func init() {
	rules = make(map[tokentype.TokenType]ParseRule)
	rules[tokentype.TOKEN_LEFT_PAREN] = ParseRule{(*Parser).grouping, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_RIGHT_PAREN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_LEFT_BRACE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_RIGHT_BRACE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_COMMA] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_DOT] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_MINUS] = ParseRule{(*Parser).unary, (*Parser).binary, precedence.PREC_TERM}
	rules[tokentype.TOKEN_PLUS] = ParseRule{nil, (*Parser).binary, precedence.PREC_TERM}
	rules[tokentype.TOKEN_SEMICOLON] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_SLASH] = ParseRule{nil, (*Parser).binary, precedence.PREC_FACTOR}
	rules[tokentype.TOKEN_STAR] = ParseRule{nil, (*Parser).binary, precedence.PREC_FACTOR}
	rules[tokentype.TOKEN_BANG] = ParseRule{(*Parser).unary, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_BANG_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_EQUALITY}
	rules[tokentype.TOKEN_EQUAL] = ParseRule{nil, nil, precedence.PREC_EQUALITY}
	rules[tokentype.TOKEN_EQUAL_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_GREATER] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_GREATER_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_LESS] = ParseRule{nil, (*Parser).binary, precedence.PREC_COMPARISON}
	rules[tokentype.TOKEN_LESS_EQUAL] = ParseRule{nil, (*Parser).binary, precedence.PREC_NONE}
	rules[tokentype.TOKEN_IDENTIFIER] = ParseRule{(*Parser).variable, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_STRING] = ParseRule{(*Parser).string, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_NUMBER] = ParseRule{(*Parser).number, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_AND] = ParseRule{nil, (*Parser).and_, precedence.PREC_NONE}
	rules[tokentype.TOKEN_CLASS] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_ELSE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FALSE] = ParseRule{(*Parser).literal, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FOR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FUN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_IF] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_NIL] = ParseRule{(*Parser).literal, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_OR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_PRINT] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_RETURN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_SUPER] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_THIS] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_TRUE] = ParseRule{(*Parser).literal, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_VAR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_WHILE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_ERROR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_EOF] = ParseRule{nil, nil, precedence.PREC_NONE}
}

// Compile the input source string and emits byteCode
func Compile(source string, chunk *chunk.Chunk) bool {
	scanner := scanner.New()
	scanner.InitScanner(source)
	compiler := newCompiler()

	parser := New(scanner, chunk, compiler)

	parser.advance()

	for !parser.match(tokentype.TOKEN_EOF) {
		parser.declaration()
	}

	parser.endCompiler()

	return !parser.HadError
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

	offset := len(parser.currentChunk().Code) - loopStart + 2
	// if (offset > UINT16_MAX) error("Loop body too large.");

	parser.emitByte(byte((offset >> 8) & 0xff))
	parser.emitByte(byte(offset & 0xff))
}

func (parser *Parser) emitJump(instruction opcode.OpCode) int {
	parser.emitByte(byte(instruction))
	parser.emitByte(0xff)
	parser.emitByte(0xff)
	return len(parser.currentChunk().Code) - 2
}

func (parser *Parser) emitReturn() {
	parser.emitByte(byte(opcode.OP_RETURN))
}

func (parser *Parser) emitConstant(value value.Value) {
	parser.currentChunk().WriteConstant(value, parser.Previous.Line)
}

func (parser *Parser) patchJump(offset int) {
	jump := len(parser.currentChunk().Code) - offset - 2

	// if (jump > UINT16_MAX) {
	// 	error("Too much code to jump over.");
	//   }

	parser.currentChunk().Code[offset] = byte((jump >> 8) & 0xff)
	parser.currentChunk().Code[offset+1] = byte(jump & 0xff)
}

func (parser *Parser) endCompiler() {
	parser.emitReturn()

	if config.DEBUG_PRINT_CODE {
		if !parser.HadError {
			debug.DisassembleChunk(parser.currentChunk(), "code")
		}
	}
}

func (parser *Parser) beginScope() {
	parser.CurrentC.ScopeDepth++
}

func (parser *Parser) endScope() {
	parser.CurrentC.ScopeDepth--

	for len(parser.CurrentC.Locals) > 0 && parser.CurrentC.Locals[len(parser.CurrentC.Locals)-1].depth > parser.CurrentC.ScopeDepth {
		parser.emitByte(byte(opcode.OP_POP))
		parser.CurrentC.Locals = parser.CurrentC.Locals[:len(parser.CurrentC.Locals)-1]
	}
}

func (parser *Parser) binary(canAssign bool) {
	// Remember the operator
	operatorType := parser.Previous.Type

	// Compile the right operand
	rule := parser.getRule(operatorType)
	parser.parsePrecedence(precedence.Precedence(rule.Precedence + 1))

	switch operatorType {
	case tokentype.TOKEN_BANG_EQUAL:
		parser.emitBytes(byte(opcode.OP_EQUAL), byte(opcode.OP_NOT))
		break
	case tokentype.TOKEN_EQUAL_EQUAL:
		parser.emitByte(byte(opcode.OP_EQUAL))
		break
	case tokentype.TOKEN_GREATER:
		parser.emitByte(byte(opcode.OP_GREATER))
		break
	case tokentype.TOKEN_GREATER_EQUAL:
		parser.emitBytes(byte(opcode.OP_LESS), byte(opcode.OP_NOT))
		break
	case tokentype.TOKEN_LESS:
		parser.emitByte(byte(opcode.OP_LESS))
		break
	case tokentype.TOKEN_LESS_EQUAL:
		parser.emitBytes(byte(opcode.OP_GREATER), byte(opcode.OP_NOT))
		break
	case tokentype.TOKEN_PLUS:
		parser.emitByte(byte(opcode.OP_ADD))
		break
	case tokentype.TOKEN_MINUS:
		parser.emitBytes(byte(opcode.OP_NEGATE), byte(opcode.OP_ADD))
		break
	case tokentype.TOKEN_STAR:
		parser.emitByte(byte(opcode.OP_MULTIPLY))
		break
	case tokentype.TOKEN_SLASH:
		parser.emitByte(byte(opcode.OP_DIVIDE))
		break
	default:
		return
	}
}

func (parser *Parser) literal(canAssign bool) {
	switch parser.Previous.Type {
	case tokentype.TOKEN_FALSE:
		parser.emitByte(byte(opcode.OP_FALSE))
		break
	case tokentype.TOKEN_NIL:
		parser.emitByte(byte(opcode.OP_NIL))
		break
	case tokentype.TOKEN_TRUE:
		parser.emitByte(byte(opcode.OP_TRUE))
		break
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

func (parser *Parser) or_(canAssign bool) {
	elseJump := parser.emitJump(opcode.OP_JUMP_IF_FALSE)
	endJump := parser.emitJump(opcode.OP_JUMP)

	parser.patchJump(elseJump)
	parser.emitByte(byte(opcode.OP_POP))

	parser.parsePrecedence(precedence.PREC_OR)
	parser.patchJump(endJump)
}

func (parser *Parser) string(canAssign bool) {
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
		if arg < 256 {
			parser.emitBytes(byte(setOp), byte(arg))
		} else {
			parser.emitByte(byte(setOpLong))
			parser.emitByte(byte(arg & 0xff))
			parser.emitByte(byte((arg >> 8) & 0xff))
			parser.emitByte(byte((arg >> 16) & 0xff))
		}
	} else {
		if arg < 256 {
			parser.emitBytes(byte(getOp), byte(arg))
		} else {
			parser.emitByte(byte(getOpLong))
			parser.emitByte(byte(arg & 0xff))
			parser.emitByte(byte((arg >> 8) & 0xff))
			parser.emitByte(byte((arg >> 16) & 0xff))
		}
	}
}

func (parser *Parser) variable(canAssign bool) {
	parser.namedVariable(parser.Previous, canAssign)
}

func (parser *Parser) unary(canAssign bool) {
	operatorType := parser.Previous.Type

	// Compile the operand
	parser.parsePrecedence(precedence.PREC_UNARY)

	switch operatorType {
	case tokentype.TOKEN_BANG:
		parser.emitByte(byte(opcode.OP_NOT))
		break
	case tokentype.TOKEN_MINUS:
		parser.emitByte(byte(opcode.OP_NEGATE))
		break
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
	for i := len(parser.CurrentC.Locals) - 1; i >= 0; i-- {
		local := &parser.CurrentC.Locals[i]
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
	parser.CurrentC.Locals = append(parser.CurrentC.Locals, local)
}

func (parser *Parser) declareVariable() {
	// Global variables are implicitly declared.
	if parser.CurrentC.ScopeDepth == 0 {
		return
	}

	name := &parser.Previous
	for i := len(parser.CurrentC.Locals) - 1; i >= 0; i-- {
		local := &parser.CurrentC.Locals[i]
		if local.depth != -1 && local.depth < parser.CurrentC.ScopeDepth {
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
	if parser.CurrentC.ScopeDepth > 0 {
		return 0
	}

	return parser.identifierConstant(&parser.Previous)
}

func (parser *Parser) markInitialized() {
	parser.CurrentC.Locals[len(parser.CurrentC.Locals)-1].depth = parser.CurrentC.ScopeDepth
}

func (parser *Parser) defineVariable(global int) {
	if parser.CurrentC.ScopeDepth > 0 {
		parser.markInitialized()
		return
	}

	if global < 256 {
		parser.emitBytes(byte(opcode.OP_DEFINE_GLOBAL), byte(global))
	} else {
		parser.emitByte(byte(opcode.OP_DEFINE_GLOBAL_LONG))
		parser.emitByte(byte(global & 0xff))
		parser.emitByte(byte((global >> 8) & 0xff))
		parser.emitByte(byte((global >> 16) & 0xff))
	}
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

func (parser *Parser) whileStatement() {
	loopStart := len(parser.currentChunk().Code)

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
	if parser.match(tokentype.TOKEN_VAR) {
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
	} else if parser.match(tokentype.TOKEN_IF) {
		parser.ifStatement()
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
	return parser.chunk
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
