package vm

import (
	"golanglox/lib/chunk"
	"testing"
)

func createChunkForTesting(bytes ...byte) *chunk.Chunk {
	c := chunk.New()
	for _, b := range bytes {
		c.WriteChunk(b, 1)
	}
	return c
}

func TestReadByte(t *testing.T) {
	// test for reading the current chunk byte
	t.Run("read current chunk byte", func(t *testing.T) {
		v := New()
		v.InitVM()

		c := createChunkForTesting(1, 2, 3)

		v.Chunk = c
		v.IP = &(v.Chunk.Code[0])

		if v.readByte() != 1 {
			t.Errorf("vm.readByte() failed, expected to read the current byte")
		}
	})

	// test for incrementing the instruction pointer(IP)
	t.Run("increment instruction pointer(IP)", func(t *testing.T) {
		v := New()
		v.InitVM()

		c := createChunkForTesting(1, 2, 3)

		v.Chunk = c
		v.IP = &(v.Chunk.Code[0])

		v.readByte()
		if v.readByte() != 2 {
			t.Errorf("vm.readByte() failed, expected to increment instruction pointer(IP)")
		}
	})
}
