package opcode

type OpCode byte

const (
	OP_CONSTANT byte = iota
	OP_CONSTANT_LONG
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NEGATE
	OP_RETURN
)
