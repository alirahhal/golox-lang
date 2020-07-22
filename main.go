package main

import (
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/vm"
)

func main() {
	vm := vm.New()
	vm.InitVM()
	chunk := chunk.New()

	chunk.WriteConstant(1.2, 123)
	chunk.WriteConstant(5, 123)
	chunk.WriteChunk(opcode.OP_ADD, 123)
	chunk.WriteConstant(6.2, 123)
	chunk.WriteChunk(opcode.OP_SUBTRACT, 123)
	chunk.WriteChunk(opcode.OP_RETURN, 123)

	vm.Interpret(chunk)

	vm.FreeVM()
	chunk.FreeChunk()
}
