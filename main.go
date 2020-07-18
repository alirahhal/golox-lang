package main

import (
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/debug"
)

func main() {
	var chunk chunk.Chunk

	chunk.WriteConstant(1.2, 123)
	chunk.WriteConstant(2, 123)
	chunk.WriteConstant(5, 123)
	chunk.WriteConstant(700.2, 123)
	chunk.WriteChunk(opcode.OP_RETURN, 123)
	debug.DisassembleChunk(&chunk, "test chunk")
}
