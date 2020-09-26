package chunk

import (
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/value"
)

// Chunk struct
type Chunk struct {
	Code      []byte
	Lines     []int
	Constants value.ValueArray
}

// New return a pointer to a new Chunk struct
func New() *Chunk {
	return new(Chunk)
}

// WriteChunk append a new byte code to chunk
func (chunk *Chunk) WriteChunk(b byte, line int) {
	chunk.Code = append(chunk.Code, b)
	chunk.Lines = append(chunk.Lines, line)
}

// WriteConstant write the required bytecode for storing the value in constants
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

// FreeChunk reset the chunk
func (chunk *Chunk) FreeChunk() {
	chunk.Code = make([]byte, 0)
	chunk.Lines = make([]int, 0)
	chunk.Constants.FreeValueArray()
}

// AddConstant appends a new value to the constants slice and returns its index
func (chunk *Chunk) AddConstant(value value.Value) int {
	chunk.Constants.WriteValueArray(value)
	return len(chunk.Constants.Values) - 1
}
