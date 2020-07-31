package compiler

import (
	"fmt"
	"golanglox/lib/scanner"
	"golanglox/lib/scanner/tokentype"
)

func Compile(source string) {
	scanner := scanner.New()
	scanner.InitScanner(source)

	line := -1
	for {
		token := scanner.ScanToken()

		if token.Type == tokentype.TOKEN_EOF {
			break
		}

		if token.Line != line {
			fmt.Printf("%4d ", token.Line)
			line = token.Line
		} else {
			fmt.Print("   | ")
		}
		fmt.Printf("%2d '%s'\n", token.Type, token.Lexeme)

	}
}
