package chunk

import (
	"golox-lang/lib/chunk/opcode"
	"golox-lang/lib/value"
)

type Chunk struct {
	code      []byte
	lines     []int
	constants value.ValueArray
}

func New() *Chunk {
	return new(Chunk)
}

func (chunk *Chunk) GetCode() []byte {
	return chunk.code
}

func (chunk *Chunk) GetLines() []int {
	return chunk.lines
}

func (chunk *Chunk) GetConstants() value.ValueArray {
	return chunk.constants
}

func (chunk *Chunk) WriteChunk(b byte, line int) {
	chunk.code = append(chunk.code, b)
	chunk.lines = append(chunk.lines, line)
}

func (chunk *Chunk) WriteConstant(value value.Value, line int) {
	index := chunk.AddConstant(value)

	if index < 256 {
		chunk.WriteChunk(byte(opcode.OP_CONSTANT), line)
		chunk.WriteChunk(byte(index), line)
	} else {
		chunk.WriteChunk(byte(opcode.OP_CONSTANT_LONG), line)
		chunk.WriteChunk(byte(index&0xff), line)
		chunk.WriteChunk(byte((index>>8)&0xff), line)
		chunk.WriteChunk(byte((index>>16)&0xff), line)
	}
}

func (chunk *Chunk) FreeChunk() {
	chunk.code = make([]byte, 0)
	chunk.lines = make([]int, 0)
	chunk.constants.FreeValueArray()
}

func (chunk *Chunk) AddConstant(value value.Value) int {
	chunk.constants.WriteValueArray(value)
	return len(chunk.constants.Values) - 1
}
