package opcode

type OpCode byte

const (
	OP_CONSTANT byte = iota
	OP_CONSTANT_LONG
	OP_RETURN
)
