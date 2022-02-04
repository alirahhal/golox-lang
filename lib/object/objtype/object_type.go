package objtype

type ObjectType byte

const (
	OBJ_FUNCTION ObjectType = iota
	OBJ_NATIVE
	OBJ_STRING
	OBJ_CLASS
	OBJ_INSTANCE
	OBJ_BOUND_METHOD
)
