package chunk

type Chunk struct {
	Code []byte
}

func (chunk *Chunk) FreeChunk() {
	chunk.Code = nil
}

func (chunk *Chunk) WriteChunk(b byte) {
	chunk.Code = append(chunk.Code, b)
}
