package compiler

import (
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/compiler/precedence"
	"golanglox/lib/config"
	"golanglox/lib/debug"
	"golanglox/lib/object/objecttype"
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

	scanner *scanner.Scanner // should be a global variable ???
	chunk   *chunk.Chunk     // should be a global variable ???
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
func New(scanner *scanner.Scanner, chunk *chunk.Chunk) *Parser {
	parser := new(Parser)
	parser.HadError = false
	parser.PanicMode = false
	parser.scanner = scanner
	parser.chunk = chunk
	return parser
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
	rules[tokentype.TOKEN_AND] = ParseRule{nil, nil, precedence.PREC_NONE}
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

	parser := New(scanner, chunk)

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

func (parser *Parser) emitReturn() {
	parser.emitByte(byte(opcode.OP_RETURN))
}

func (parser *Parser) emitConstant(value value.Value) {
	parser.currentChunk().WriteConstant(value, parser.Previous.Line)
}

func (parser *Parser) endCompiler() {
	parser.emitReturn()

	if config.DEBUG_PRINT_CODE {
		if !parser.HadError {
			debug.DisassembleChunk(parser.currentChunk(), "code")
		}
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

func (parser *Parser) string(canAssign bool) {
	parser.emitConstant(
		value.NewObjString(
			&value.ObjString{Obj: value.Obj{Type: objecttype.OBJ_STRING}, String: parser.Previous.Lexeme[1 : len(parser.Previous.Lexeme)-1]}))
}

func (parser *Parser) namedVariable(name token.Token, canAssign bool) {
	arg := parser.identifierConstant(&name)
	
	if canAssign && parser.match(tokentype.TOKEN_EQUAL) {
		parser.expression()
		if arg < 256 {
			parser.emitBytes(byte(opcode.OP_SET_GLOBAL), byte(arg))
		} else {
			parser.emitByte(byte(opcode.OP_SET_GLOBAL_LONG))
			parser.emitByte(byte(arg&0xff))
			parser.emitByte(byte((arg>>8)&0xff))
			parser.emitByte(byte((arg>>16)&0xff))
		}
	} else {
		if arg < 256 {
			parser.emitBytes(byte(opcode.OP_GET_GLOBAL), byte(arg))
		} else {
			parser.emitByte(byte(opcode.OP_GET_GLOBAL_LONG))
			parser.emitByte(byte(arg&0xff))
			parser.emitByte(byte((arg>>8)&0xff))
			parser.emitByte(byte((arg>>16)&0xff))
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
	return parser.currentChunk().AddConstant(value.NewObjString(
		&value.ObjString{Obj: value.Obj{Type: objecttype.OBJ_STRING}, String: name.Lexeme}))
}

func (parser *Parser) parserVariable(errorMessage string) int {
	parser.consume(tokentype.TOKEN_IDENTIFIER, errorMessage)
	return parser.identifierConstant(&parser.Previous)
}

func (parser *Parser) defineVariable(global int) {
	if global < 256 {
		parser.emitBytes(byte(opcode.OP_DEFINE_GLOBAL), byte(global))
	} else {
		parser.emitByte(byte(opcode.OP_DEFINE_GLOBAL_LONG))
		parser.emitByte(byte(global&0xff))
		parser.emitByte(byte((global>>8)&0xff))
		parser.emitByte(byte((global>>16)&0xff))
	}
}

func (parser *Parser) getRule(token tokentype.TokenType) *ParseRule {
	rule := rules[token]
	return &rule
}

func (parser *Parser) expression() {
	parser.parsePrecedence(precedence.PREC_ASSIGNMENT)
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

func (parser *Parser) printStatement() {
	parser.expression()
	parser.consume(tokentype.TOKEN_SEMICOLON, "Expect ';' after value.")
	parser.emitByte(byte(opcode.OP_PRINT))
}

func (parser *Parser) synchronize() {
	parser.PanicMode = false
	for parser.Current.Type != tokentype.TOKEN_EOF {
		if(parser.Previous.Type == tokentype.TOKEN_SEMICOLON) {
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
		  return;
  
		default:
		  // Do nothing.
		  ;
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
	if (parser.match(tokentype.TOKEN_PRINT)) {
		parser.printStatement()
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
