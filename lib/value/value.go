package value

import (
	"fmt"
	"golox-lang/lib/object"
	"golox-lang/lib/object/objtype"
	"golox-lang/lib/value/valuetype"
	"unsafe"
)

type FuncChunk interface {
	GetCode() []byte
	GetLines() []int
	GetConstants() ValueArray
}

type ObjFunction struct {
	object.Obj
	Arity int
	Chunk FuncChunk
	Name  *ObjString
}

type NativeFn func(argCount int, args []Value) Value

type ObjNative struct {
	object.Obj
	Function NativeFn
}

type ObjString struct {
	object.Obj
	String string
}

type ObjClass struct {
	object.Obj
	Name string
}

type ObjInstance struct {
	object.Obj
	Klass  *ObjClass
	Fields map[string]Value
}

type Value struct {
	Type valuetype.ValueType
	Data interface{}
}

func New(valType valuetype.ValueType, val interface{}) Value {
	return Value{Type: valType, Data: val}
}

func NewFunction(c FuncChunk) *ObjFunction {
	valObj := &ObjFunction{Obj: object.Obj{Type: objtype.OBJ_FUNCTION}, Arity: 0, Name: nil, Chunk: c}
	return valObj
}

func NewObjFunction(val *ObjFunction) Value {
	return Value{Type: valuetype.VAL_OBJ, Data: (*object.Obj)(unsafe.Pointer(val))}
}

func NewNative(function NativeFn) *ObjNative {
	native := &ObjNative{Obj: object.Obj{Type: objtype.OBJ_NATIVE}, Function: function}
	return native
}

func NewObjNative(val *ObjNative) Value {
	return Value{Type: valuetype.VAL_OBJ, Data: (*object.Obj)(unsafe.Pointer(val))}
}

func NewObjString(val string) Value {
	valObj := &ObjString{Obj: object.Obj{Type: objtype.OBJ_STRING}, String: val}
	return Value{Type: valuetype.VAL_OBJ, Data: (*object.Obj)(unsafe.Pointer(valObj))}
}

func NewObjClass(val string) Value {
	valObj := &ObjClass{Obj: object.Obj{Type: objtype.OBJ_CLASS}, Name: val}
	return Value{Type: valuetype.VAL_OBJ, Data: (*object.Obj)(unsafe.Pointer(valObj))}
}

func NewObjInstance(klass *ObjClass) Value {
	valObj := &ObjInstance{Obj: object.Obj{Type: objtype.OBJ_INSTANCE}, Klass: klass, Fields: make(map[string]Value)}
	return Value{Type: valuetype.VAL_OBJ, Data: (*object.Obj)(unsafe.Pointer(valObj))}
}

func (value Value) AsBool() bool {
	return value.Data.(bool)
}

func (value Value) AsNumber() float64 {
	return value.Data.(float64)
}

func (value Value) AsObj() *object.Obj {
	return value.Data.(*object.Obj)
}

func (value Value) AsFunction() *ObjFunction {
	return (*ObjFunction)(unsafe.Pointer(value.AsObj()))
}

func (value Value) AsNative() *ObjNative {
	return (*ObjNative)(unsafe.Pointer(value.AsObj()))
}

func (value Value) AsString() *ObjString {
	return (*ObjString)(unsafe.Pointer(value.AsObj()))
}

func (value Value) AsClass() *ObjClass {
	return (*ObjClass)(unsafe.Pointer(value.AsObj()))
}

func (value Value) AsInstance() *ObjInstance {
	return (*ObjInstance)(unsafe.Pointer(value.AsObj()))
}

func (value Value) AsGoString() string {
	return value.AsString().String
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

func (value Value) ObjType() objtype.ObjectType {
	return value.AsObj().Type
}

func (value Value) IsFunction() bool {
	return value.isobjtype(objtype.OBJ_FUNCTION)
}

func (value Value) IsNative() bool {
	return value.isobjtype(objtype.OBJ_NATIVE)
}

func (value Value) IsString() bool {
	return value.isobjtype(objtype.OBJ_STRING)
}

func (value Value) IsClass() bool {
	return value.isobjtype(objtype.OBJ_CLASS)
}

func (value Value) IsInstance() bool {
	return value.isobjtype(objtype.OBJ_INSTANCE)
}

func (value Value) isobjtype(objectType objtype.ObjectType) bool {
	return value.IsObj() && value.ObjType() == objectType
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

	case valuetype.VAL_NIL:
		fmt.Printf("nil")

	case valuetype.VAL_NUMBER:
		fmt.Printf("%g", value.AsNumber())

	case valuetype.VAL_OBJ:
		value.PrintObject()

	}
}

func printFunction(function *ObjFunction) {
	if function.Name == nil {
		fmt.Printf("<script>")
		return
	}
	fmt.Printf("<fn %s>", function.Name.String)
}

func (value Value) PrintObject() {
	switch value.ObjType() {
	case objtype.OBJ_FUNCTION:
		printFunction(value.AsFunction())

	case objtype.OBJ_NATIVE:
		fmt.Printf("<native fn>")

	case objtype.OBJ_STRING:
		fmt.Printf("%s", value.AsGoString())

	case objtype.OBJ_CLASS:
		fmt.Printf("%s", value.AsClass().Name)

	case objtype.OBJ_INSTANCE:
		fmt.Printf("%s instance", value.AsInstance().Klass.Name)

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
		aObj := a.AsObj()
		bObj := b.AsObj()

		if aObj.Type == objtype.OBJ_STRING {
			return a.AsGoString() == b.AsGoString()
		} else {
			return aObj == bObj
		}
	default:
		return false
	}
}
