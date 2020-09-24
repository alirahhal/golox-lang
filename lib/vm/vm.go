package vm

import (
	"encoding/binary"
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/compiler"
	"golanglox/lib/config"
	"golanglox/lib/debug"
	"golanglox/lib/object/objecttype"
	"golanglox/lib/unsafecode"
	"golanglox/lib/value"
	"golanglox/lib/value/valuetype"
	"golanglox/lib/vm/interpretresult"
	"os"
)

const (
	STACK_INITIAL_SIZE int = 256
)

// VM struct
type VM struct {
	Chunk *chunk.Chunk
	IP    *byte
	Stack []value.Value
	Globals map[string]value.Value
}

// New return a pointer to a new VM struct
func New() *VM {
	vm := new(VM)
	return vm
}

// InitVM intitialze VM struct
func (vm *VM) InitVM() {
	vm.resetStack()
}

// Interpret takes a source code and fires off the VM execution pipeline
func (vm *VM) Interpret(source string) interpretresult.InterpretResult {
	chunk := chunk.New()

	if !compiler.Compile(source, chunk) {
		chunk.FreeChunk()
		return interpretresult.INTERPRET_COMPILE_ERROR
	}

	vm.Chunk = chunk
	vm.IP = &(vm.Chunk.Code[0])

	result := vm.run()

	chunk.FreeChunk()
	return result
}

func (vm *VM) FreeVM() {}

func (vm *VM) run() interpretresult.InterpretResult {
	for {
		if config.DEBUG_TRACE_EXECUTION {
			fmt.Print("          ")
			for _, val := range vm.Stack {
				fmt.Print("[ ")
				val.PrintValue()
				fmt.Print(" ]")
			}
			fmt.Print("\n")
			debug.DisassembleInstruction(vm.Chunk, unsafecode.Diff(vm.IP, &(vm.Chunk.Code[0])))
		}

		var instruction opcode.OpCode
		switch instruction = opcode.OpCode(vm.readByte()); instruction {
		case opcode.OP_CONSTANT:
			var constant value.Value = vm.readConstant()
			vm.push(constant)
			break
		case opcode.OP_CONSTANT_LONG:
			var constant value.Value = vm.readConstantLong()
			vm.push(constant)
			break
		case opcode.OP_NIL:
			vm.push(value.New(valuetype.VAL_NIL, nil))
			break
		case opcode.OP_TRUE:
			vm.push(value.New(valuetype.VAL_BOOL, true))
			break
		case opcode.OP_FALSE:
			vm.push(value.New(valuetype.VAL_BOOL, false))
			break
		case opcode.OP_POP:
			vm.pop()
			break
		case opcode.OP_GET_GLOBAL:
			name := vm.readConstant().AsGoString()
			val, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.push(val)
			break
		case opcode.OP_GET_GLOBAL_LONG:
			name := vm.readConstantLong().AsGoString()
			val, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.push(val)
			break
		case opcode.OP_DEFINE_GLOBAL:
			name := vm.readConstant().AsGoString()
			vm.Globals[name] = vm.peek(0)
			vm.pop()
			break
		case opcode.OP_DEFINE_GLOBAL_LONG:
			name := vm.readConstantLong().AsGoString()
			vm.Globals[name] = vm.peek(0)
			vm.pop()
			break
		case opcode.OP_SET_GLOBAL:
			name := vm.readConstant().AsGoString()
			_, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.Globals[name] = vm.peek(0)
			break
		case opcode.OP_SET_GLOBAL_LONG:
			name := vm.readConstantLong().AsGoString()
			_, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.Globals[name] = vm.peek(0)
			break
		case opcode.OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(value.New(valuetype.VAL_BOOL, value.ValuesEqual(a, b)))
			break
		case opcode.OP_GREATER:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_BOOL, func(a, b value.Value) interface{} {
				return a.AsNumber() > b.AsNumber()
			})
			break
		case opcode.OP_LESS:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_BOOL, func(a, b value.Value) interface{} {
				return a.AsNumber() < b.AsNumber()
			})
			break
		
		case opcode.OP_ADD:
			if vm.peek(0).IsString() && vm.peek(1).IsString() {
				vm.concatenate()
			} else if vm.peek(0).IsNumber() && vm.peek(1).IsNumber() {
				vm.binaryOP(valuetype.VAL_NUMBER, func(a, b value.Value) interface{} {
					return a.AsNumber() + b.AsNumber()
				})
			} else {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			break
		case opcode.OP_MULTIPLY:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_NUMBER, func(a, b value.Value) interface{} {
				return a.AsNumber() * b.AsNumber()
			})
			break
		case opcode.OP_DIVIDE:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_NUMBER, func(a, b value.Value) interface{} {
				return a.AsNumber() / b.AsNumber()
			})
			break
		case opcode.OP_NOT:
			vm.push(value.New(valuetype.VAL_BOOL, isFalsey(vm.pop())))
			break
		case opcode.OP_NEGATE:
			if !vm.peek(0).IsNumber() {
				vm.runtimeError("Operand must be a number")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.push(value.New(valuetype.VAL_NUMBER, -vm.pop().AsNumber()))
			break
		case opcode.OP_PRINT:
			vm.pop().PrintValue()
			fmt.Printf("\n")
			break
		case opcode.OP_RETURN:
			// Exit interpreter
			return interpretresult.INTERPRET_OK
		}
	}
}

// Push appends a new Value to the vm`s stack
func (vm *VM) push(val value.Value) {
	vm.Stack = append(vm.Stack, val)
}

// Pop pops a Value from the vm`s stack
func (vm *VM) pop() value.Value {
	defer vm.shrinkStack()

	var x value.Value
	x, vm.Stack = vm.Stack[len(vm.Stack)-1], vm.Stack[:len(vm.Stack)-1]
	return x
}

func (vm *VM) shrinkStack() {
	if cap(vm.Stack) > STACK_INITIAL_SIZE*2 && len(vm.Stack) <= (cap(vm.Stack)/2) {
		vm.Stack = append([]value.Value(nil), vm.Stack[:len(vm.Stack)]...)
	}
}

func (vm *VM) peek(distance int) value.Value {
	return vm.Stack[len(vm.Stack)-1-distance]
}

func isFalsey(val value.Value) bool {
	return val.IsNil() || (val.IsBool() && !val.AsBool())
}

func (vm *VM) concatenate() {
	b := vm.pop().AsGoString()
	a := vm.pop().AsGoString()

	vm.push(value.NewObjString(
		&value.ObjString{Obj: value.Obj{Type: objecttype.OBJ_STRING}, String: a + b}))
}

func (vm *VM) readByte() byte {
	returnVal := *(vm.IP)
	vm.IP = unsafecode.Increment(vm.IP)

	return returnVal
}

func (vm *VM) readConstant() value.Value {
	return vm.Chunk.Constants.Values[vm.readByte()]
}

func (vm *VM) readConstantLong() value.Value {
	constBytes := make([]byte, 0)
	for i := 0; i < 4; i++ {
		if i == 3 {
			constBytes = append(constBytes, 0)
			break
		}
		constBytes = append(constBytes, vm.readByte())
	}
	var constantAddress uint32 = binary.LittleEndian.Uint32(constBytes)
	return vm.Chunk.Constants.Values[constantAddress]
}

func (vm *VM) binaryOP(valueType valuetype.ValueType, op func(a, b value.Value) interface{}) {
	b := vm.pop()
	a := vm.pop()
	vm.push(value.New(valueType, op(a, b)))
}

func (vm *VM) resetStack() {
	vm.Stack = make([]value.Value, 0, STACK_INITIAL_SIZE)
	vm.Globals = make(map[string]value.Value)
}

func (vm *VM) runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Println()

	offset := unsafecode.Diff(vm.IP, &(vm.Chunk.Code[0]))
	line := vm.Chunk.Lines[offset]
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)

	vm.resetStack()
}
