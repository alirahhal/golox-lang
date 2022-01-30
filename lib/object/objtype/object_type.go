package objtype

type ObjectType byte

const (
	OBJ_FUNCTION ObjectType = iota
	OBJ_NATIVE
	OBJ_STRING
)
