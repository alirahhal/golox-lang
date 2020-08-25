package compiler

import (
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/compiler/precedence"
	"golanglox/lib/scanner"
	"golanglox/lib/scanner/tokentype"
	"golanglox/lib/value"
	"os"
	"strconv"
)

// Parser struct
type Parser struct {
	Current   scanner.Token
	Previous  scanner.Token
	HadError  bool
	PanicMode bool

	scanner *scanner.Scanner
	chunk   *chunk.Chunk // should be a global variable ???
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
	// expression()
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
}

func (parser *Parser) binary() {
	// Remember the operator
	operatorType := parser.Previous.Type

	// Compile the right operand
	/************** Continue Here ************/
	// var *ParseRule rule = parser.getRule(operatorType)
	// parser.parsePrecedence()

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

}

func (parser *Parser) expression() {
	parser.parsePrecedence(precedence.PREC_ASSIGNMENT)
}

func (parser *Parser) currentChunk() *chunk.Chunk {
	return parser.chunk
}

func (parser *Parser) errorAt(token *scanner.Token, message string) {
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
