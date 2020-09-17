package value

import (
	"fmt"
	"golanglox/lib/value/valuetype"
)

type Value struct {
	Type valuetype.ValueType
	Data interface{}
}

func New(valType valuetype.ValueType, val interface{}) Value {
	return Value{Type: valType, Data: val}
}

func (value Value) AsBool() bool {
	return value.Data.(bool)
}

func (value Value) AsNumber() float64 {
	return value.Data.(float64)
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
	fmt.Printf("%g", value.AsNumber())
}
