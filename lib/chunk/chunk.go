package chunk

import (
	"golanglox/lib/value"
)

type Chunk struct {
	Code      []byte
	Lines     []int
	Constants value.ValueArray
}

func (chunk *Chunk) WriteChunk(b byte, line int) {
	chunk.Code = append(chunk.Code, b)
	chunk.Lines = append(chunk.Lines, line)
}

func (chunk *Chunk) FreeChunk() {
	chunk.Code = nil
	chunk.Lines = nil
	chunk.Constants.FreeValueArray()
}

func (chunk *Chunk) AddConstant(value value.Value) int {
	chunk.Constants.WriteValueArray(value)
	return len(chunk.Constants.Values) - 1
}
