package chunk

import (
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/value"
)

type Chunk struct {
	Code      []byte
	Lines     []int
	Constants value.ValueArray
}

func New() *Chunk {
	chunk := new(Chunk)
	chunk.Code = make([]byte, 0)
	chunk.Lines = make([]int, 0)
	return chunk
}

func (chunk *Chunk) WriteChunk(b byte, line int) {
	chunk.Code = append(chunk.Code, b)
	chunk.Lines = append(chunk.Lines, line)
}

func (chunk *Chunk) WriteConstant(value value.Value, line int) {
	index := chunk.addConstant(value)

	if index < 256 {
		chunk.WriteChunk(opcode.OP_CONSTANT, line)
		chunk.WriteChunk(byte(index), line)
	} else {
		chunk.WriteChunk(opcode.OP_CONSTANT_LONG, line)
		chunk.WriteChunk(byte(index&0xff), line)
		chunk.WriteChunk(byte((index>>8)&0xff), line)
		chunk.WriteChunk(byte((index>>16)&0xff), line)
	}
}

func (chunk *Chunk) FreeChunk() {
	chunk.Code = nil
	chunk.Lines = nil
	chunk.Constants.FreeValueArray()
}

func (chunk *Chunk) addConstant(value value.Value) int {
	chunk.Constants.WriteValueArray(value)
	return len(chunk.Constants.Values) - 1
}
