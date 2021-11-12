package vm

import (
	"golox-lang/lib/chunk"
)

func createChunkForTesting(bytes ...byte) *chunk.Chunk {
	c := chunk.New()
	for _, b := range bytes {
		c.WriteChunk(b, 1)
	}
	return c
}
