package scanner_test

import (
	"golanglox/lib/scanner"
	"golanglox/lib/scanner/token/tokentype"
	"testing"
)

func TestScanToken(t *testing.T) {
	// test for skipping leading whitespaces
	t.Run("skip leading whitespaces", func(t *testing.T) {
		s := scanner.New()
		s.InitScanner(" \r\t hello world")

		s.ScanToken()
		if s.Source[s.Start:] != "hello world" {
			t.Errorf("chunk.ScanToken() failed, expected to remove leading whitespaces")
		}
	})

	// test for incementing line
	t.Run("increment line", func(t *testing.T) {
		s := scanner.New()
		s.InitScanner(" \n hello world")

		s.ScanToken()
		if s.Line != 2 {
			t.Errorf("chunk.ScanToken() failed, expected to increment line")
		}
	})

	// test for all kind of tokens
	t.Run("test different tokens", func(t *testing.T) {
		dataItems := []struct {
			source          string
			wantedTokenType tokentype.TokenType
			wantedLexeme    string
		}{
			{
				source:          "{",
				wantedTokenType: tokentype.TOKEN_LEFT_BRACE,
				wantedLexeme:    "{",
			},
			{
				source:          "and",
				wantedTokenType: tokentype.TOKEN_AND,
				wantedLexeme:    "and",
			},
			{
				source:          "class",
				wantedTokenType: tokentype.TOKEN_CLASS,
				wantedLexeme:    "class",
			},
			{
				source:          "!",
				wantedTokenType: tokentype.TOKEN_BANG,
				wantedLexeme:    "!",
			},
			{
				source:          "!=",
				wantedTokenType: tokentype.TOKEN_BANG_EQUAL,
				wantedLexeme:    "!=",
			},
			{
				source:          ",",
				wantedTokenType: tokentype.TOKEN_COMMA,
				wantedLexeme:    ",",
			},
			{
				source:          ".",
				wantedTokenType: tokentype.TOKEN_DOT,
				wantedLexeme:    ".",
			},
			{
				source:          "else",
				wantedTokenType: tokentype.TOKEN_ELSE,
				wantedLexeme:    "else",
			},
			{
				source:          "",
				wantedTokenType: tokentype.TOKEN_EOF,
				wantedLexeme:    "",
			},
			{
				source:          "=",
				wantedTokenType: tokentype.TOKEN_EQUAL,
				wantedLexeme:    "=",
			},
			{
				source:          "==",
				wantedTokenType: tokentype.TOKEN_EQUAL_EQUAL,
				wantedLexeme:    "==",
			},
			{
				source:          "\"abc",
				wantedTokenType: tokentype.TOKEN_ERROR,
				wantedLexeme:    "Unterminated string.",
			},
			{
				source:          "false",
				wantedTokenType: tokentype.TOKEN_FALSE,
				wantedLexeme:    "false",
			},
			{
				source:          "for",
				wantedTokenType: tokentype.TOKEN_FOR,
				wantedLexeme:    "for",
			},
			{
				source:          "fun",
				wantedTokenType: tokentype.TOKEN_FUN,
				wantedLexeme:    "fun",
			},
			{
				source:          ">",
				wantedTokenType: tokentype.TOKEN_GREATER,
				wantedLexeme:    ">",
			},
			{
				source:          ">=",
				wantedTokenType: tokentype.TOKEN_GREATER_EQUAL,
				wantedLexeme:    ">=",
			},
			{
				source:          "id",
				wantedTokenType: tokentype.TOKEN_IDENTIFIER,
				wantedLexeme:    "id",
			},
			{
				source:          "if",
				wantedTokenType: tokentype.TOKEN_IF,
				wantedLexeme:    "if",
			},
			{
				source:          "{",
				wantedTokenType: tokentype.TOKEN_LEFT_BRACE,
				wantedLexeme:    "{",
			},
			{
				source:          "(",
				wantedTokenType: tokentype.TOKEN_LEFT_PAREN,
				wantedLexeme:    "(",
			},
			{
				source:          "<",
				wantedTokenType: tokentype.TOKEN_LESS,
				wantedLexeme:    "<",
			},
			{
				source:          "<=",
				wantedTokenType: tokentype.TOKEN_LESS_EQUAL,
				wantedLexeme:    "<=",
			},
			{
				source:          "-",
				wantedTokenType: tokentype.TOKEN_MINUS,
				wantedLexeme:    "-",
			},
			{
				source:          "nil",
				wantedTokenType: tokentype.TOKEN_NIL,
				wantedLexeme:    "nil",
			},
			{
				source:          "123.2",
				wantedTokenType: tokentype.TOKEN_NUMBER,
				wantedLexeme:    "123.2",
			},
			{
				source:          "or",
				wantedTokenType: tokentype.TOKEN_OR,
				wantedLexeme:    "or",
			},
			{
				source:          "+",
				wantedTokenType: tokentype.TOKEN_PLUS,
				wantedLexeme:    "+",
			},
			{
				source:          "print",
				wantedTokenType: tokentype.TOKEN_PRINT,
				wantedLexeme:    "print",
			},
			{
				source:          "return",
				wantedTokenType: tokentype.TOKEN_RETURN,
				wantedLexeme:    "return",
			},
			{
				source:          "}",
				wantedTokenType: tokentype.TOKEN_RIGHT_BRACE,
				wantedLexeme:    "}",
			},
			{
				source:          ")",
				wantedTokenType: tokentype.TOKEN_RIGHT_PAREN,
				wantedLexeme:    ")",
			},
			{
				source:          ";",
				wantedTokenType: tokentype.TOKEN_SEMICOLON,
				wantedLexeme:    ";",
			},
			{
				source:          "/",
				wantedTokenType: tokentype.TOKEN_SLASH,
				wantedLexeme:    "/",
			},
			{
				source:          "*",
				wantedTokenType: tokentype.TOKEN_STAR,
				wantedLexeme:    "*",
			},
			{
				source:          "\"hellow world\"",
				wantedTokenType: tokentype.TOKEN_STRING,
				wantedLexeme:    "\"hellow world\"",
			},
			{
				source:          "super",
				wantedTokenType: tokentype.TOKEN_SUPER,
				wantedLexeme:    "super",
			},
			{
				source:          "this",
				wantedTokenType: tokentype.TOKEN_THIS,
				wantedLexeme:    "this",
			},
			{
				source:          "true",
				wantedTokenType: tokentype.TOKEN_TRUE,
				wantedLexeme:    "true",
			},
			{
				source:          "var",
				wantedTokenType: tokentype.TOKEN_VAR,
				wantedLexeme:    "var",
			},
			{
				source:          "while",
				wantedTokenType: tokentype.TOKEN_WHILE,
				wantedLexeme:    "while",
			},
		}

		for _, item := range dataItems {
			s := scanner.New()
			s.InitScanner(item.source)
			tkn := s.ScanToken()

			if tkn.Type != item.wantedTokenType || tkn.Lexeme != item.wantedLexeme {
				t.Errorf("chunk.ScanToken() failed, excpected to get token %v of type %v, got token %v of type %v",
					item.wantedLexeme, item.wantedTokenType, tkn.Lexeme, tkn.Type)
			}
		}
	})
}
