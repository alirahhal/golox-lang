package token

import "golanglox/lib/scanner/token/tokentype"

type Token struct {
	Type   tokentype.TokenType
	Lexeme string
	Line   int
}

func MakeToken(tokenType tokentype.TokenType, lexem string, line int) Token {
	var token Token
	token.Type = tokenType
	token.Lexeme = lexem
	token.Line = line

	return token
}
