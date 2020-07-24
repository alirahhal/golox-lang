package main

import (
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/vm"
	"log"
	"time"
)

func main() {
	start := time.Now()

	vm := vm.New()
	vm.InitVM()
	chunk := chunk.New()

	chunk.WriteConstant(1, 123)
	chunk.WriteConstant(2, 123)
	chunk.WriteConstant(3, 123)
	chunk.WriteConstant(4, 123)
	chunk.WriteConstant(5, 123)
	chunk.WriteConstant(6, 123)
	chunk.WriteConstant(7, 123)
	chunk.WriteChunk(opcode.OP_ADD, 123)
	// chunk.WriteChunk(opcode.OP_SUBTRACT, 123)
	chunk.WriteChunk(opcode.OP_RETURN, 123)

	// debug.DisassembleChunk(chunk, "test")

	vm.Interpret(chunk)

	vm.FreeVM()
	chunk.FreeChunk()

	elapsed := time.Since(start)
	fmt.Print("\n\n")
	log.Printf("Running took %s", elapsed)
}
