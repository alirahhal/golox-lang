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
	"os"
	"strconv"
)

// Parser struct
type Parser struct {
	Current   token.Token
	Previous  token.Token
	HadError  bool
	PanicMode bool

	scanner *scanner.Scanner
	chunk   *chunk.Chunk // should be a global variable ???
}

// ParseFn func
type ParseFn func()

// ParseRule struct
type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence precedence.Precedence
}

// New creates a new parser and returns it
func New(scanner *scanner.Scanner, chunk *chunk.Chunk) *Parser {
	parser := new(Parser)
	parser.HadError = false
	parser.PanicMode = false
	parser.scanner = scanner
	parser.chunk = chunk
	return parser
}

// Compile the input source string and emits byteCode
func Compile(source string, chunk *chunk.Chunk) bool {
	scanner := scanner.New()
	scanner.InitScanner(source)

	parser := New(scanner, chunk)

	parser.advance()
	parser.expression()
	parser.consume(tokentype.TOKEN_EOF, "Expect end of expression.")
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

func (parser *Parser) emitByte(b byte) {
	parser.currentChunk().WriteChunk(b, parser.Previous.Line)
}

func (parser *Parser) emitBytes(b1 byte, b2 byte) {
	parser.emitByte(b1)
	parser.emitByte(b2)
}

func (parser *Parser) emitReturn() {
	parser.emitByte(opcode.OP_RETURN)
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

func (parser *Parser) binary() {
	// Remember the operator
	operatorType := parser.Previous.Type

	// Compile the right operand
	rule := parser.getRule(operatorType)
	parser.parsePrecedence(precedence.Precedence(rule.Precedence + 1))

	switch operatorType {
	case tokentype.TOKEN_PLUS:
		parser.emitByte(opcode.OP_ADD)
		break
	case tokentype.TOKEN_MINUS:
		parser.emitByte(opcode.OP_SUBTRACT)
		break
	case tokentype.TOKEN_STAR:
		parser.emitByte(opcode.OP_MULTIPLY)
		break
	case tokentype.TOKEN_SLASH:
		parser.emitByte(opcode.OP_DIVIDE)
		break
	default:
		return
	}
}

func (parser *Parser) grouping() {
	parser.expression()
	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (parser *Parser) number() {
	val, err := strconv.ParseFloat(parser.Previous.Lexeme, 64)
	if err != nil {
	}
	parser.emitConstant(value.Value(val))
}

func (parser *Parser) unary() {
	operatorType := parser.Previous.Type

	// Compile the operand
	parser.parsePrecedence(precedence.PREC_UNARY)

	switch operatorType {
	case tokentype.TOKEN_MINUS:
		parser.emitByte(opcode.OP_NEGATE)
		break
	default:
		return
	}
}

func (parser *Parser) parsePrecedence(precedence precedence.Precedence) {
	parser.advance()
	prefixRule := parser.getRule(parser.Previous.Type).Prefix
	if prefixRule == nil {
		parser.error("Expect Expression.")
		return
	}

	prefixRule()

	for precedence <= parser.getRule(parser.Current.Type).Precedence {
		parser.advance()
		infixRule := parser.getRule(parser.Previous.Type).Infix
		infixRule()
	}
}

func (parser *Parser) getRule(token tokentype.TokenType) *ParseRule {
	rules := map[tokentype.TokenType]ParseRule{
		tokentype.TOKEN_LEFT_PAREN:    {parser.grouping, nil, precedence.PREC_NONE},
		tokentype.TOKEN_RIGHT_PAREN:   {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_LEFT_BRACE:    {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_RIGHT_BRACE:   {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_COMMA:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_DOT:           {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_MINUS:         {parser.unary, parser.binary, precedence.PREC_TERM},
		tokentype.TOKEN_PLUS:          {nil, parser.binary, precedence.PREC_TERM},
		tokentype.TOKEN_SEMICOLON:     {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_SLASH:         {nil, parser.binary, precedence.PREC_FACTOR},
		tokentype.TOKEN_STAR:          {nil, parser.binary, precedence.PREC_FACTOR},
		tokentype.TOKEN_BANG:          {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_BANG_EQUAL:    {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_EQUAL:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_EQUAL_EQUAL:   {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_GREATER:       {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_GREATER_EQUAL: {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_LESS:          {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_LESS_EQUAL:    {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_IDENTIFIER:    {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_STRING:        {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_NUMBER:        {parser.number, nil, precedence.PREC_NONE},
		tokentype.TOKEN_AND:           {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_CLASS:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_ELSE:          {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_FALSE:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_FOR:           {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_FUN:           {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_IF:            {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_NIL:           {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_OR:            {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_PRINT:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_RETURN:        {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_SUPER:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_THIS:          {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_TRUE:          {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_VAR:           {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_WHILE:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_ERROR:         {nil, nil, precedence.PREC_NONE},
		tokentype.TOKEN_EOF:           {nil, nil, precedence.PREC_NONE},
	}

	rule := rules[token]
	return &rule
}

func (parser *Parser) expression() {
	parser.parsePrecedence(precedence.PREC_ASSIGNMENT)
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
