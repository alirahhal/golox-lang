package tokentype

type TokenType byte

const (
	// Single-character tokens.
	TOKEN_LEFT_PAREN  TokenType = iota // 1 - 1
	TOKEN_RIGHT_PAREN                  // 2 - 1
	TOKEN_LEFT_BRACE                   // 3 - 1
	TOKEN_RIGHT_BRACE                  // 4 - 1
	TOKEN_COMMA                        // 5 - 1
	TOKEN_DOT                          // 6 - 1
	TOKEN_MINUS                        // 7 - 1
	TOKEN_PLUS                         // 8 - 1
	TOKEN_SEMICOLON                    // 9 - 1
	TOKEN_SLASH                        // 10 - 1
	TOKEN_STAR                         // 11 - 1

	// One or two character tokens.
	TOKEN_BANG          // 12 - 1
	TOKEN_BANG_EQUAL    // 13 - 1
	TOKEN_EQUAL         // 14 - 1
	TOKEN_EQUAL_EQUAL   // 15 - 1
	TOKEN_GREATER       // 16 - 1
	TOKEN_GREATER_EQUAL // 17 - 1
	TOKEN_LESS          // 18 - 1
	TOKEN_LESS_EQUAL    // 19 - 1

	// Literals.
	TOKEN_IDENTIFIER // 20 - 1
	TOKEN_STRING     // 21 - 1
	TOKEN_NUMBER     // 22 - 1

	// Keywords.
	TOKEN_AND    // 23 - 1
	TOKEN_CLASS  // 24 - 1
	TOKEN_ELSE   // 25 - 1
	TOKEN_FALSE  // 26 - 1
	TOKEN_FOR    // 27 - 1
	TOKEN_FUN    // 28 - 1
	TOKEN_IF     // 29 - 1
	TOKEN_NIL    // 30 - 1
	TOKEN_OR     // 31 - 1
	TOKEN_PRINT  // 32 - 1
	TOKEN_RETURN // 33 - 1
	TOKEN_SUPER  // 34 - 1
	TOKEN_THIS   // 35 - 1
	TOKEN_TRUE   // 36 - 1
	TOKEN_VAR    // 37 - 1
	TOKEN_WHILE  // 38 - 1

	TOKEN_ERROR // 39 - 1
	TOKEN_EOF   // 40 - 1
)
