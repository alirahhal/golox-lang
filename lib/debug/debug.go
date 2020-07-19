package debug

import (
	"encoding/binary"
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
	if offset > 0 && chunk.Lines[offset] == chunk.Lines[offset-1] {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%4d ", chunk.Lines[offset])
	}

	instruction := chunk.Code[offset]
	switch instruction {
	case opcode.OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", chunk, offset)
	case opcode.OP_CONSTANT_LONG:
		return longConstantInstruction("OP_CONSTANT_LONG", chunk, offset)
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

func constantInstruction(name string, chunk *chunk.Chunk, offset int) int {
	constant := chunk.Code[offset+1]
	fmt.Printf("%s %4d ", name, constant)
	chunk.Constants.Values[constant].PrintValue()
	fmt.Print("\n")
	return offset + 2
}

func longConstantInstruction(name string, chunk *chunk.Chunk, offset int) int {
	constByte := make([]byte, 4)
	copy(constByte, chunk.Code[offset+1:offset+4])
	var constant uint32 = binary.LittleEndian.Uint32(constByte)
	fmt.Printf("%s %4d ", name, constant)
	chunk.Constants.Values[constant].PrintValue()
	fmt.Print("\n")
	return offset + 4
}