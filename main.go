package main

import (
	"bufio"
	"fmt"
	"golox-lang/lib/vm"
	"golox-lang/lib/vm/interpretresult"
	"io/ioutil"

	"os"
)

func main() {

	vm := vm.New()
	vm.InitVM()

	if len(os.Args) == 1 {
		repl(vm)
	} else if len(os.Args) == 2 {
		runFile(os.Args[1], vm)
	} else {
		fmt.Print("Wrong number of arguments\n")
		os.Exit(64)
	}

	vm.FreeVM()
}

func repl(vm *vm.VM) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		line, err := reader.ReadString(byte(13))

		if err != nil {
			fmt.Print("\n")
			break
		}
		vm.Interpret(line)
	}
}

func runFile(path string, vm *vm.VM) {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	source := string(fileContent)

	result := vm.Interpret(source)

	if result == interpretresult.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if result == interpretresult.INTERPRET_RUNTIME_ERROR {
		os.Exit(70)
	}
}
