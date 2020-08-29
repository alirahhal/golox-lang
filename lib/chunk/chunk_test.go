package chunk

import (
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/value"
	"testing"
)

func TestWriteChunk(t *testing.T) {
	chunk := New()
	var b byte = 10
	var l = 1
	chunk.WriteChunk(b, l)

	if bA, lA := chunk.Code[len(chunk.Code)-1], chunk.Lines[len(chunk.Lines)-1]; bA != b || lA != l {
		t.Errorf("chunk.WriteChunk(%v, %v) failed, expected to append %v to Code, and %v to Line, got %v, %v respectively", b, l, b, l, bA, lA)
	}
}

func TestWriteConstant(t *testing.T) {
	// test for constant less than 256 elements
	t.Run("index<256", func(t *testing.T) {
		chunk := New()

		chunk.WriteConstant(10, 0)
		if op := opcode.OpCode(chunk.Code[len(chunk.Code)-2]); op != opcode.OP_CONSTANT {
			t.Errorf("chunk.WriteConstant(...) failed, expected to get OpCode %v, got %v", opcode.OP_CONSTANT, op)
		}
	})

	// test for constant greater than 256 elements
	t.Run("index>=256", func(t *testing.T) {
		chunk := New()

		for i := 0; i < 257; i++ {
			chunk.WriteConstant(value.Value(i), 0)
		}
		if op := opcode.OpCode(chunk.Code[len(chunk.Code)-4]); op != opcode.OP_CONSTANT_LONG {
			t.Errorf("chunk.WriteConstant(...) failed, expected to get OpCode %v, got %v", opcode.OP_CONSTANT_LONG, op)
		}
	})
}
