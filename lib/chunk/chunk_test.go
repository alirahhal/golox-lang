package chunk

import (
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/value"
	"testing"
)

func TestWriteChunk(t *testing.T) {
	chunkCreated := New()
	var byteToAppend byte = 10
	var line = 1
	chunkCreated.WriteChunk(byteToAppend, line)

	if bA, lA := chunkCreated.Code[len(chunkCreated.Code)-1], chunkCreated.Lines[len(chunkCreated.Lines)-1]; bA != byteToAppend || lA != line {
		t.Errorf("chunk.WriteChunk(%v, %v) failed, expected to append %v to Code, and %v to Line, got %v, %v respectively",
			byteToAppend, line, byteToAppend, line, bA, lA)
	}
}

func TestWriteConstant(t *testing.T) {
	// test for constant less than 256 elements
	t.Run("index<256", func(t *testing.T) {
		chunkCreated := New()

		chunkCreated.WriteConstant(10, 0)
		if op := opcode.OpCode(chunkCreated.Code[len(chunkCreated.Code)-2]); op != opcode.OP_CONSTANT {
			t.Errorf("chunk.WriteConstant(...) failed, expected to get OpCode %v, got %v", opcode.OP_CONSTANT, op)
		}
	})

	// test for constant greater than 256 elements
	t.Run("index>=256", func(t *testing.T) {
		chunkCreated := New()

		for i := 0; i < 257; i++ {
			chunkCreated.WriteConstant(value.Value(i), 0)
		}
		if op := opcode.OpCode(chunkCreated.Code[len(chunkCreated.Code)-4]); op != opcode.OP_CONSTANT_LONG {
			t.Errorf("chunk.WriteConstant(...) failed, expected to get OpCode %v, got %v", opcode.OP_CONSTANT_LONG, op)
		}
	})
}
