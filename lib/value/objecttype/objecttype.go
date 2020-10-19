package objecttype

type ObjType byte

const (
	OBJ_FUNCTION ObjType = iota
	OBJ_NATIVE
	OBJ_STRING
)
