package value

import (
	"fmt"
	"golanglox/lib/object/objecttype"
	"golanglox/lib/value/valuetype"
	"unsafe"
)

type Obj struct {
	Type objecttype.ObjType
}
type ObjString struct {
	Obj
	String string
}

type Value struct {
	Type valuetype.ValueType
	Data interface{}
}

func New(valType valuetype.ValueType, val interface{}) Value {
	return Value{Type: valType, Data: val}
}

func NewObjString(val *ObjString) Value {
	return Value{Type: valuetype.VAL_OBJ, Data: (*Obj)(unsafe.Pointer(val))}
}

func (value Value) AsBool() bool {
	return value.Data.(bool)
}

func (value Value) AsNumber() float64 {
	return value.Data.(float64)
}

func (value Value) AsObj() *Obj {
	return value.Data.(*Obj)
}

func (value Value) IsBool() bool {
	return value.Type == valuetype.VAL_BOOL
}

func (value Value) IsNil() bool {
	return value.Type == valuetype.VAL_NIL
}

func (value Value) IsNumber() bool {
	return value.Type == valuetype.VAL_NUMBER
}

func (value Value) IsObj() bool {
	return value.Type == valuetype.VAL_OBJ
}

func (value Value) ObjType() objecttype.ObjType {
	return value.AsObj().Type
}

func (value Value) IsString() bool {
	return value.isObjType(objecttype.OBJ_STRING)
}

func (value Value) AsString() *ObjString {
	return (*ObjString)(unsafe.Pointer(value.AsObj()))
}

func (value Value) AsGoString() string {
	return ((*ObjString)(unsafe.Pointer(value.AsObj()))).String
}

func (value Value) isObjType(objType objecttype.ObjType) bool {
	return value.IsObj() && value.ObjType() == objType
}

type ValueArray struct {
	Values []Value
}

func (valueArray *ValueArray) WriteValueArray(value Value) {
	valueArray.Values = append(valueArray.Values, value)
}

func (valueArray *ValueArray) FreeValueArray() {
	valueArray.Values = make([]Value, 0)
}

func (value Value) PrintValue() {
	switch value.Type {
	case valuetype.VAL_BOOL:
		if value.AsBool() {
			fmt.Printf("true")
		} else {
			fmt.Printf("false")
		}
		break
	case valuetype.VAL_NIL:
		fmt.Printf("nil")
		break
	case valuetype.VAL_NUMBER:
		fmt.Printf("%g", value.AsNumber())
		break
	case valuetype.VAL_OBJ:
		value.PrintObject()
		break
	}
}

func (value Value) PrintObject() {
	switch value.ObjType() {
	case objecttype.OBJ_STRING:
		fmt.Printf("%s", value.AsGoString())
		break
	}
}

func ValuesEqual(a Value, b Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case valuetype.VAL_BOOL:
		return a.AsBool() == b.AsBool()
	case valuetype.VAL_NIL:
		return true
	case valuetype.VAL_NUMBER:
		return a.AsNumber() == b.AsNumber()
	case valuetype.VAL_OBJ:
		aString := a.AsGoString()
		bString := b.AsGoString()
		return aString == bString
	default:
		return false
	}
}
