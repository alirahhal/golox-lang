package main

import (
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/debug"
)

func main() {
	var chunk chunk.Chunk

	constant := chunk.AddConstant(1.2)
	chunk.WriteChunk(opcode.OP_CONSTANT, 123)
	chunk.WriteChunk(byte(constant), 123)
	chunk.WriteChunk(opcode.OP_RETURN, 123)
	debug.DisassembleChunk(&chunk, "test chunk")
}
