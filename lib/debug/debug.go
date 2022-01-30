package debug

import (
	"encoding/binary"
	"fmt"
	"golox-lang/lib/chunk"
	"golox-lang/lib/chunk/opcode"
)

func DisassembleChunk(chunk *chunk.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(chunk.GetCode()); {
		offset = DisassembleInstruction(chunk, offset)
	}
}

func DisassembleInstruction(chunk *chunk.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)
	if offset > 0 && chunk.GetLines()[offset] == chunk.GetLines()[offset-1] {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%4d ", chunk.GetLines()[offset])
	}

	instruction := opcode.OpCode(chunk.GetCode()[offset])
	switch instruction {
	case opcode.OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", chunk, offset)
	case opcode.OP_CONSTANT_LONG:
		return longConstantInstruction("OP_CONSTANT_LONG", chunk, offset)
	case opcode.OP_NIL:
		return simpleInstruction("OP_NIL", offset)
	case opcode.OP_TRUE:
		return simpleInstruction("OP_TRUE", offset)
	case opcode.OP_FALSE:
		return simpleInstruction("OP_FALSE", offset)
	case opcode.OP_POP:
		return simpleInstruction("OP_POP", offset)
	case opcode.OP_GET_LOCAL:
		return byteInstruction("OP_GET_LOCAL", chunk, offset)
	case opcode.OP_GET_LOCAL_LONG:
		return byteInstructionLong("OP_GET_LOCAL_LONG", chunk, offset)
	case opcode.OP_SET_LOCAL:
		return byteInstruction("OP_SET_LOCAL", chunk, offset)
	case opcode.OP_SET_LOCAL_LONG:
		return byteInstructionLong("OP_SET_LOCAL_LONG", chunk, offset)
	case opcode.OP_GET_GLOBAL:
		return constantInstruction("OP_GET_GLOBAL", chunk, offset)
	case opcode.OP_GET_GLOBAL_LONG:
		return longConstantInstruction("OP_GET_GLOBAL_LONG", chunk, offset)
	case opcode.OP_DEFINE_GLOBAL:
		return constantInstruction("OP_DEFINE_GLOBAL", chunk, offset)
	case opcode.OP_DEFINE_GLOBAL_LONG:
		return longConstantInstruction("OP_DEFINE_GLOBAL_LONG", chunk, offset)
	case opcode.OP_SET_GLOBAL:
		return constantInstruction("OP_SET_GLOBAL", chunk, offset)
	case opcode.OP_SET_GLOBAL_LONG:
		return longConstantInstruction("OP_SET_GLOBAL_LONG", chunk, offset)
	case opcode.OP_EQUAL:
		return simpleInstruction("OP_EQUAL", offset)
	case opcode.OP_GREATER:
		return simpleInstruction("OP_GREATER", offset)
	case opcode.OP_LESS:
		return simpleInstruction("OP_LESS", offset)
	case opcode.OP_ADD:
		return simpleInstruction("OP_ADD", offset)
	case opcode.OP_MULTIPLY:
		return simpleInstruction("OP_MULTIPLY", offset)
	case opcode.OP_DIVIDE:
		return simpleInstruction("OP_DIVIDE", offset)
	case opcode.OP_NOT:
		return simpleInstruction("OP_NOT", offset)
	case opcode.OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case opcode.OP_PRINT:
		return simpleInstruction("OP_PRINT", offset)
	case opcode.OP_JUMP:
		return jumpInstruction("OP_JUMP", 1, chunk, offset)
	case opcode.OP_JUMP_IF_FALSE:
		return jumpInstruction("OP_JUMP_IF_FALSE", 1, chunk, offset)
	case opcode.OP_LOOP:
		return jumpInstruction("OP_LOOP", -1, chunk, offset)
	case opcode.OP_CALL:
		return byteInstruction("OP_CALL", chunk, offset)
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

func byteInstruction(name string, chunk *chunk.Chunk, offset int) int {
	slot := chunk.GetCode()[offset+1]
	fmt.Printf("%s %4d\n", name, slot)
	return offset + 2
}

func byteInstructionLong(name string, chunk *chunk.Chunk, offset int) int {
	bytes := make([]byte, 4)
	copy(bytes, chunk.GetCode()[offset+1:offset+4])
	var slot uint32 = binary.LittleEndian.Uint32(bytes)
	fmt.Printf("%s %4d\n", name, slot)
	return offset + 4
}

func jumpInstruction(name string, sign int, chunk *chunk.Chunk, offset int) int {
	bytes := make([]byte, 4)
	copy(bytes, chunk.GetCode()[offset+1:offset+3])
	jump := binary.BigEndian.Uint16(bytes)
	fmt.Printf("%s %4d -> %d\n", name, offset, offset+3+sign*int(jump))
	return offset + 3
}

func constantInstruction(name string, chunk *chunk.Chunk, offset int) int {
	constant := chunk.GetCode()[offset+1]
	fmt.Printf("%s %4d ", name, constant)
	chunk.GetConstants().Values[constant].PrintValue()
	fmt.Print("\n")
	return offset + 2
}

func longConstantInstruction(name string, chunk *chunk.Chunk, offset int) int {
	constBytes := make([]byte, 4)
	copy(constBytes, chunk.GetCode()[offset+1:offset+4])
	var constant uint32 = binary.LittleEndian.Uint32(constBytes)
	fmt.Printf("%s %4d ", name, constant)
	chunk.GetConstants().Values[constant].PrintValue()
	fmt.Print("\n")
	return offset + 4
}
