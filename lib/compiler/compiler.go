package compiler

import (
	"golanglox/lib/scanner"
)

func Compile(source string) {
	scanner := scanner.New()
	scanner.InitScanner(source)
}
