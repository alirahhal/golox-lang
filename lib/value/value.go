package value

import (
	"fmt"
)

// maybe Value should be an interface, and make others types implement it
type Value float64

type ValueArray struct {
	Values []Value
}

func (valueArray *ValueArray) WriteValueArray(value Value) {
	valueArray.Values = append(valueArray.Values, value)
}

func (valueArray *ValueArray) FreeValueArray() {
	valueArray.Values = nil
}

func (value *Value) PrintValue() {
	fmt.Printf("%g", float64(*value))
}
