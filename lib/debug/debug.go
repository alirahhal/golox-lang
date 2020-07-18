package debug

import (
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
)

func DisassembleChunk(chunk *chunk.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(chunk.Code); {
		offset = DisassembleInstruction(chunk, offset)
	}
}

func DisassembleInstruction(chunk *chunk.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)

	instruction := chunk.Code[offset]
	switch instruction {
	case opcode.OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Println("Unknown opcode ", instruction)
		return offset + 1
	}
}

func simpleInstruction(name string, offset int) int {
	fmt.Println(name)
	return offset + 1
}
