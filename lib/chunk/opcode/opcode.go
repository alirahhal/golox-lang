package opcode

type OpCode byte

const (
	OP_CONSTANT OpCode = iota
	OP_CONSTANT_LONG
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NEGATE
	OP_RETURN
)
