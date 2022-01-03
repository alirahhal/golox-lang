package chunk

import (
	"golox-lang/lib/chunk/opcode"
	"golox-lang/lib/value"
)

type Chunk struct {
	Code      []byte
	Lines     []int
	Constants value.ValueArray
}

func New() *Chunk {
	return new(Chunk)
}

func (chunk *Chunk) GetCode() []byte {
	return chunk.Code
}

func (chunk *Chunk) GetLines() []int {
	return chunk.Lines
}

func (chunk *Chunk) GetConstants() value.ValueArray {
	return chunk.Constants
}

func (chunk *Chunk) WriteChunk(b byte, line int) {
	chunk.Code = append(chunk.Code, b)
	chunk.Lines = append(chunk.Lines, line)
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
	chunk.Code = make([]byte, 0)
	chunk.Lines = make([]int, 0)
	chunk.Constants.FreeValueArray()
}

func (chunk *Chunk) AddConstant(value value.Value) int {
	chunk.Constants.WriteValueArray(value)
	return len(chunk.Constants.Values) - 1
}
