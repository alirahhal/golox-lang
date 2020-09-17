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

	scanner *scanner.Scanner // should be a global variable ???
	chunk   *chunk.Chunk     // should be a global variable ???
}

// ParseFn func
type ParseFn func(receiver *Parser)

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
	rules[tokentype.TOKEN_BANG] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_BANG_EQUAL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_EQUAL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_EQUAL_EQUAL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_GREATER] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_GREATER_EQUAL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_LESS] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_LESS_EQUAL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_IDENTIFIER] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_STRING] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_NUMBER] = ParseRule{(*Parser).number, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_AND] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_CLASS] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_ELSE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FALSE] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FOR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_FUN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_IF] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_NIL] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_OR] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_PRINT] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_RETURN] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_SUPER] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_THIS] = ParseRule{nil, nil, precedence.PREC_NONE}
	rules[tokentype.TOKEN_TRUE] = ParseRule{nil, nil, precedence.PREC_NONE}
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

func (parser *Parser) binary() {
	// Remember the operator
	operatorType := parser.Previous.Type

	// Compile the right operand
	rule := parser.getRule(operatorType)
	parser.parsePrecedence(precedence.Precedence(rule.Precedence + 1))

	switch operatorType {
	case tokentype.TOKEN_PLUS:
		parser.emitByte(byte(opcode.OP_ADD))
		break
	case tokentype.TOKEN_MINUS:
		parser.emitByte(byte(opcode.OP_SUBTRACT))
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

func (parser *Parser) grouping() {
	parser.expression()
	parser.consume(tokentype.TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (parser *Parser) number() {
	val, err := strconv.ParseFloat(parser.Previous.Lexeme, 64)
	if err != nil {
	}
	parser.emitConstant(value.New(valuetype.VAL_NUMBER, val))
}

func (parser *Parser) unary() {
	operatorType := parser.Previous.Type

	// Compile the operand
	parser.parsePrecedence(precedence.PREC_UNARY)

	switch operatorType {
	case tokentype.TOKEN_MINUS:
		parser.emitByte(byte(opcode.OP_NEGATE))
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

	prefixRule(parser)

	for precedence <= parser.getRule(parser.Current.Type).Precedence {
		parser.advance()
		infixRule := parser.getRule(parser.Previous.Type).Infix
		// Not sure!!!
		if infixRule == nil {
			parser.error("Expect Expression.")
			return
		}
		//
		infixRule(parser)
	}
}

func (parser *Parser) getRule(token tokentype.TokenType) *ParseRule {
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
