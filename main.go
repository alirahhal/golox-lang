package main

import (
	"bufio"
	"fmt"
	"golanglox/lib/vm"
	"golanglox/lib/vm/interpretresult"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()

	vm := vm.New()
	vm.InitVM()

	if len(os.Args) == 1 {
		repl()
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		fmt.Print("Usage: clox [path]\n")
		os.Exit(64)
	}

	vm.FreeVM()

	elapsed := time.Since(start)
	fmt.Print("\n\n")
	log.Printf("Running took %s", elapsed)
}

func repl() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		line, err := reader.ReadString('\n')

		if err != nil {
			fmt.Print("\n")
			break
		}
		interpret(line)
	}
}

func runFile(path string) {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	source := string(fileContent)
	interpretresult.InterpretResult result = interpret(source)

	ifresult == interpretresult.INTERPRET_COMPILE_ERROR {os.Exit(65)}
	ifresult == interpretresult.INTERPRET_COMPILE_ERROR {os.Exit(70)}
}
