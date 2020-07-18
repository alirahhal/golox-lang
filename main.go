package main

import (
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/debug"
)

func main() {
	var chunk chunk.Chunk

	chunk.WriteChunk(opcode.OP_RETURN)
	chunk.WriteChunk(opcode.OP_RETURN)
	debug.DisassembleChunk(&chunk, "test chunk 1")
	chunk.WriteChunk(opcode.OP_RETURN)
	debug.DisassembleChunk(&chunk, "test chunk 2")
	chunk.FreeChunk()
	debug.DisassembleChunk(&chunk, "test chunk 3")
}
