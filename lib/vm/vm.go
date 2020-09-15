package vm

import (
	"encoding/binary"
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/compiler"
	"golanglox/lib/config"
	"golanglox/lib/debug"
	"golanglox/lib/unsafecode"
	"golanglox/lib/value"
	"golanglox/lib/vm/interpretresult"
	"unsafe"
)

const (
	STACK_INITIAL_SIZE int = 256
)

// VM struct
type VM struct {
	Chunk *chunk.Chunk
	IP    *byte
	Stack []value.Value
}

// New return a pointer to a new VM struct
func New() *VM {
	vm := new(VM)
	// vm.Stack = make([]value.Value, 0, STACK_INITIAL_SIZE)
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
			debug.DisassembleInstruction(vm.Chunk, int(uintptr(unsafe.Pointer(vm.IP))-uintptr(unsafe.Pointer(&(vm.Chunk.Code[0])))))
		}

		var instruction byte
		switch instruction = vm.readByte(); instruction {
		case byte(opcode.OP_CONSTANT):
			var constant value.Value = vm.readConstant()
			vm.push(constant)
			break
		case byte(opcode.OP_CONSTANT_LONG):
			var constant value.Value = vm.readConstantLong()
			vm.push(constant)
			break
		case byte(opcode.OP_NEGATE):
			vm.push(-vm.pop())
			break
		case byte(opcode.OP_ADD):
			vm.binaryOP(func(a, b value.Value) value.Value {
				return a + b
			})
			break
		case byte(opcode.OP_SUBTRACT):
			vm.binaryOP(func(a, b value.Value) value.Value {
				return a - b
			})
			break
		case byte(opcode.OP_MULTIPLY):
			vm.binaryOP(func(a, b value.Value) value.Value {
				return a * b
			})
			break
		case byte(opcode.OP_DIVIDE):
			vm.binaryOP(func(a, b value.Value) value.Value {
				return a / b
			})
			break
		case byte(opcode.OP_RETURN):
			poped := vm.pop()
			(&poped).PrintValue()
			fmt.Print("\n")
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
	var x value.Value
	// todo: find a way for shrinking the stack based on a specific algo
	x, vm.Stack = vm.Stack[len(vm.Stack)-1], vm.Stack[:len(vm.Stack)-1]
	return x
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

func (vm *VM) binaryOP(op func(a, b value.Value) value.Value) {
	b := vm.pop()
	a := vm.pop()
	vm.push(op(a, b))
}

func (vm *VM) resetStack() {
	vm.Stack = make([]value.Value, 0, STACK_INITIAL_SIZE)
}
