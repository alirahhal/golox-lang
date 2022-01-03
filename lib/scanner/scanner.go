package scanner

import (
	"golox-lang/lib/scanner/token"
	"golox-lang/lib/scanner/token/tokentype"
	"unicode"
)

type Scanner struct {
	Source  string
	Start   int
	Current int
	Line    int
}

func New() *Scanner {
	return new(Scanner)
}

func (scanner *Scanner) InitScanner(source string) {
	scanner.Source = source
	scanner.Start = 0
	scanner.Current = 0
	scanner.Line = 1
}

func isDigit(c rune) bool {
	return unicode.IsDigit(c)
}

func isAlpha(c rune) bool {
	return unicode.IsLetter(c) || c == '_'
}

func (scanner *Scanner) ScanToken() token.Token {
	scanner.skipWhiteSpace()
	scanner.Start = scanner.Current

	if scanner.isAtEnd() {
		return scanner.makeToken(tokentype.TOKEN_EOF)
	}

	c := scanner.advance()

	if isAlpha(c) {
		return scanner.identifier()
	}
	if isDigit(c) {
		return scanner.number()
	}

	switch c {
	case '(':
		return scanner.makeToken(tokentype.TOKEN_LEFT_PAREN)
	case ')':
		return scanner.makeToken(tokentype.TOKEN_RIGHT_PAREN)
	case '{':
		return scanner.makeToken(tokentype.TOKEN_LEFT_BRACE)
	case '}':
		return scanner.makeToken(tokentype.TOKEN_RIGHT_BRACE)
	case ';':
		return scanner.makeToken(tokentype.TOKEN_SEMICOLON)
	case ',':
		return scanner.makeToken(tokentype.TOKEN_COMMA)
	case '.':
		return scanner.makeToken(tokentype.TOKEN_DOT)
	case '-':
		return scanner.makeToken(tokentype.TOKEN_MINUS)
	case '+':
		return scanner.makeToken(tokentype.TOKEN_PLUS)
	case '/':
		return scanner.makeToken(tokentype.TOKEN_SLASH)
	case '*':
		return scanner.makeToken(tokentype.TOKEN_STAR)
	case '!':
		tokenType := tokentype.TOKEN_BANG
		if scanner.match('=') {
			tokenType = tokentype.TOKEN_BANG_EQUAL
		}
		return scanner.makeToken(tokenType)
	case '=':
		tokenType := tokentype.TOKEN_EQUAL
		if scanner.match('=') {
			tokenType = tokentype.TOKEN_EQUAL_EQUAL
		}
		return scanner.makeToken(tokenType)
	case '<':
		tokenType := tokentype.TOKEN_LESS
		if scanner.match('=') {
			tokenType = tokentype.TOKEN_LESS_EQUAL
		}
		return scanner.makeToken(tokenType)
	case '>':
		tokenType := tokentype.TOKEN_GREATER
		if scanner.match('=') {
			tokenType = tokentype.TOKEN_GREATER_EQUAL
		}
		return scanner.makeToken(tokenType)
	case '"':
		return scanner.string()
	}

	return scanner.errorToken("Unexpected character.")
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.Current >= len(scanner.Source)
}

func (scanner *Scanner) advance() rune {
	scanner.Current++
	return []rune(scanner.Source)[scanner.Current-1]
}

func (scanner *Scanner) peek() rune {
	if scanner.isAtEnd() {
		// handle differently?
		return rune(0)
	}
	return []rune(scanner.Source)[scanner.Current]
}

func (scanner *Scanner) peekNext() rune {
	if scanner.Current+1 >= len(scanner.Source) {
		// handle differently?
		return rune(0)
	}
	return []rune(scanner.Source)[scanner.Current+1]
}

func (scanner *Scanner) match(expected rune) bool {
	if scanner.peek() != expected {
		return false
	}

	scanner.Current++
	return true
}

func (scanner *Scanner) makeToken(tokenType tokentype.TokenType) token.Token {
	return token.MakeToken(tokenType, scanner.Source[scanner.Start:scanner.Current], scanner.Line)
}

func (scanner *Scanner) errorToken(message string) token.Token {
	return token.MakeToken(tokentype.TOKEN_ERROR, message, scanner.Line)
}

func (scanner *Scanner) skipWhiteSpace() {
	for {
		c := scanner.peek()
		switch c {
		case ' ', '\r', '\t':
			scanner.advance()

		case '\n':
			scanner.Line++
			scanner.advance()

		case '/':
			if scanner.peekNext() == '/' {
				// A comment goes until the end of the line.
				for scanner.peek() != '\n' && !scanner.isAtEnd() {
					scanner.advance()
				}
			} else {
				return
			}

		default:
			return
		}
	}
}

func (scanner *Scanner) checkKeyword(start int, length int, rest string, tokenType tokentype.TokenType) tokentype.TokenType {
	if scanner.Current-scanner.Start == start+length && scanner.Source[scanner.Start+start:scanner.Start+start+length] == rest {
		return tokenType
	}

	return tokentype.TOKEN_IDENTIFIER
}

func (scanner *Scanner) identifierType() tokentype.TokenType {
	switch scanner.Source[scanner.Start] {
	case 'a':
		return scanner.checkKeyword(1, 2, "nd", tokentype.TOKEN_AND)
	case 'c':
		return scanner.checkKeyword(1, 4, "lass", tokentype.TOKEN_CLASS)
	case 'e':
		return scanner.checkKeyword(1, 3, "lse", tokentype.TOKEN_ELSE)
	case 'f':
		if scanner.Current-scanner.Start > 1 {
			switch scanner.Source[scanner.Start+1] {
			case 'a':
				return scanner.checkKeyword(2, 3, "lse", tokentype.TOKEN_FALSE)
			case 'o':
				return scanner.checkKeyword(2, 1, "r", tokentype.TOKEN_FOR)
			case 'u':
				return scanner.checkKeyword(2, 1, "n", tokentype.TOKEN_FUN)
			}
		}
	case 'i':
		return scanner.checkKeyword(1, 1, "f", tokentype.TOKEN_IF)
	case 'n':
		return scanner.checkKeyword(1, 2, "il", tokentype.TOKEN_NIL)
	case 'o':
		return scanner.checkKeyword(1, 1, "r", tokentype.TOKEN_OR)
	case 'p':
		return scanner.checkKeyword(1, 4, "rint", tokentype.TOKEN_PRINT)
	case 'r':
		return scanner.checkKeyword(1, 5, "eturn", tokentype.TOKEN_RETURN)
	case 's':
		return scanner.checkKeyword(1, 4, "uper", tokentype.TOKEN_SUPER)
	case 't':
		if scanner.Current-scanner.Start > 1 {
			switch scanner.Source[scanner.Start+1] {
			case 'h':
				return scanner.checkKeyword(2, 2, "is", tokentype.TOKEN_THIS)
			case 'r':
				return scanner.checkKeyword(2, 2, "ue", tokentype.TOKEN_TRUE)
			}
		}
	case 'v':
		return scanner.checkKeyword(1, 2, "ar", tokentype.TOKEN_VAR)
	case 'w':
		return scanner.checkKeyword(1, 4, "hile", tokentype.TOKEN_WHILE)
	}

	return tokentype.TOKEN_IDENTIFIER
}

func (scanner *Scanner) identifier() token.Token {
	for isAlpha(scanner.peek()) || isDigit(scanner.peek()) {
		scanner.advance()
	}

	return scanner.makeToken(scanner.identifierType())
}

func (scanner *Scanner) number() token.Token {
	for isDigit(scanner.peek()) {
		scanner.advance()
	}

	// Look for a fractional part.
	if scanner.peek() == '.' && isDigit(scanner.peekNext()) {
		// Consume the ".".
		scanner.advance()

		for isDigit(scanner.peek()) {
			scanner.advance()
		}
	}

	return scanner.makeToken(tokentype.TOKEN_NUMBER)
}

func (scanner *Scanner) string() token.Token {
	for scanner.peek() != '"' && !scanner.isAtEnd() {
		if scanner.peek() == '\n' {
			scanner.Line++
		}
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return scanner.errorToken("Unterminated string.")
	}

	// The closing quote.
	scanner.advance()
	return scanner.makeToken(tokentype.TOKEN_STRING)
}
