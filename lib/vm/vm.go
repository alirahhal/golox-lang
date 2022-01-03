package vm

import (
	"encoding/binary"
	"fmt"
	"golox-lang/lib/chunk"
	"golox-lang/lib/chunk/opcode"
	"golox-lang/lib/compiler"
	"golox-lang/lib/config"
	"golox-lang/lib/debug"
	"golox-lang/lib/utils/unsafecode"
	"golox-lang/lib/value"
	"golox-lang/lib/value/objecttype"
	"golox-lang/lib/value/valuetype"
	"golox-lang/lib/vm/interpretresult"
	"os"
	"time"
)

const (
	FRAMES_INITIAL_SIZE int = 64
	STACK_INITIAL_SIZE  int = FRAMES_INITIAL_SIZE * 256
)

type CallFrame struct {
	Function *value.ObjFunction
	IP       *byte
	Slots    int
}

type VM struct {
	Frames []CallFrame

	Stack   []value.Value
	Globals map[string]value.Value
}

func clockNative(argCount int, args []value.Value) value.Value {
	return value.New(valuetype.VAL_NUMBER, float64(time.Now().UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond))))
}

func New() *VM {
	return new(VM)
}

func (vm *VM) InitVM() {
	vm.resetStack()

	vm.defineNative("clock", clockNative)
}

func (vm *VM) Interpret(source string) interpretresult.InterpretResult {
	function := compiler.Compile(source)
	if function == nil {
		return interpretresult.INTERPRET_COMPILE_ERROR
	}

	vm.push(value.NewObjFunction(function))
	vm.callValue(value.NewObjFunction(function), 0)

	return vm.run()
}

func (vm *VM) FreeVM() {}

func (vm *VM) run() interpretresult.InterpretResult {
	var frame *CallFrame = &vm.Frames[len(vm.Frames)-1]

	for {
		if config.DEBUG_TRACE_EXECUTION {
			fmt.Print("          ")
			for _, val := range vm.Stack {
				fmt.Print("[ ")
				val.PrintValue()
				fmt.Print(" ]")
			}
			fmt.Print("\n")
			debug.DisassembleInstruction(frame.Function.Chunk.(*chunk.Chunk), unsafecode.Diff(frame.IP, &((frame.Function.Chunk.GetCode())[0])))
		}

		var instruction opcode.OpCode
		switch instruction = opcode.OpCode(vm.readByte()); instruction {
		case opcode.OP_CONSTANT:
			var constant value.Value = vm.readConstant()
			vm.push(constant)

		case opcode.OP_CONSTANT_LONG:
			var constant value.Value = vm.readConstantLong()
			vm.push(constant)

		case opcode.OP_NIL:
			vm.push(value.New(valuetype.VAL_NIL, nil))

		case opcode.OP_TRUE:
			vm.push(value.New(valuetype.VAL_BOOL, true))

		case opcode.OP_FALSE:
			vm.push(value.New(valuetype.VAL_BOOL, false))

		case opcode.OP_POP:
			vm.pop()

		case opcode.OP_GET_LOCAL:
			slot := vm.readByte()
			vm.push(vm.Stack[frame.Slots+int(slot)])

		case opcode.OP_GET_LOCAL_LONG:
			slot := vm.readLong()
			vm.push(vm.Stack[frame.Slots+int(slot)])

		case opcode.OP_SET_LOCAL:
			slot := vm.readByte()
			vm.Stack[frame.Slots+int(slot)] = vm.peek(0)

		case opcode.OP_SET_LOCAL_LONG:
			slot := vm.readLong()
			vm.Stack[frame.Slots+int(slot)] = vm.peek(0)

		case opcode.OP_GET_GLOBAL:
			name := vm.readConstant().AsGoString()
			val, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.push(val)

		case opcode.OP_GET_GLOBAL_LONG:
			name := vm.readConstantLong().AsGoString()
			val, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.push(val)

		case opcode.OP_DEFINE_GLOBAL:
			name := vm.readConstant().AsGoString()
			vm.Globals[name] = vm.peek(0)
			vm.pop()

		case opcode.OP_DEFINE_GLOBAL_LONG:
			name := vm.readConstantLong().AsGoString()
			vm.Globals[name] = vm.peek(0)
			vm.pop()

		case opcode.OP_SET_GLOBAL:
			name := vm.readConstant().AsGoString()
			_, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.Globals[name] = vm.peek(0)

		case opcode.OP_SET_GLOBAL_LONG:
			name := vm.readConstantLong().AsGoString()
			_, present := vm.Globals[name]
			if !present {
				vm.runtimeError("Undefined variable '%s'.", name)
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			vm.Globals[name] = vm.peek(0)

		case opcode.OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(value.New(valuetype.VAL_BOOL, value.ValuesEqual(a, b)))

		case opcode.OP_GREATER:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_BOOL, func(a, b value.Value) interface{} {
				return a.AsNumber() > b.AsNumber()
			})

		case opcode.OP_LESS:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_BOOL, func(a, b value.Value) interface{} {
				return a.AsNumber() < b.AsNumber()
			})

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

		case opcode.OP_MULTIPLY:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_NUMBER, func(a, b value.Value) interface{} {
				return a.AsNumber() * b.AsNumber()
			})

		case opcode.OP_DIVIDE:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				vm.runtimeError("Operands must be numbers.")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.binaryOP(valuetype.VAL_NUMBER, func(a, b value.Value) interface{} {
				return a.AsNumber() / b.AsNumber()
			})

		case opcode.OP_NOT:
			vm.push(value.New(valuetype.VAL_BOOL, isFalsey(vm.pop())))

		case opcode.OP_NEGATE:
			if !vm.peek(0).IsNumber() {
				vm.runtimeError("Operand must be a number")
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}

			vm.push(value.New(valuetype.VAL_NUMBER, -vm.pop().AsNumber()))

		case opcode.OP_PRINT:
			vm.pop().PrintValue()
			fmt.Printf("\n")

		case opcode.OP_JUMP:
			offset := vm.readShort()
			frame.IP = unsafecode.Increment(frame.IP, int(offset))

		case opcode.OP_JUMP_IF_FALSE:
			offset := vm.readShort()
			if isFalsey(vm.peek(0)) {
				frame.IP = unsafecode.Increment(frame.IP, int(offset))
			}

		case opcode.OP_LOOP:
			offset := vm.readShort()
			frame.IP = unsafecode.Decrement(frame.IP, int(offset))

		case opcode.OP_CALL:
			argCount := vm.readByte()
			if !vm.callValue(vm.peek(int(argCount)), int(argCount)) {
				return interpretresult.INTERPRET_RUNTIME_ERROR
			}
			frame = &vm.Frames[len(vm.Frames)-1]

		case opcode.OP_RETURN:
			result := vm.pop()

			vm.popFrame()
			if len(vm.Frames) == 0 {
				vm.pop()
				return interpretresult.INTERPRET_OK
			}

			for {
				if len(vm.Stack) == frame.Slots {
					break
				}
				vm.pop()
			}
			vm.push(result)

			frame = &vm.Frames[len(vm.Frames)-1]

		}
	}
}

func (vm *VM) push(val value.Value) {
	vm.Stack = append(vm.Stack, val)
}

func (vm *VM) pop() value.Value {
	defer vm.shrinkStack()

	var x value.Value
	x, vm.Stack = vm.Stack[len(vm.Stack)-1], vm.Stack[:len(vm.Stack)-1]
	return x
}

func (vm *VM) popFrame() {
	vm.Frames = vm.Frames[:len(vm.Frames)-1]
}

func (vm *VM) shrinkStack() {
	if cap(vm.Stack) > STACK_INITIAL_SIZE*2 && len(vm.Stack) <= (cap(vm.Stack)/2) {
		vm.Stack = append([]value.Value(nil), vm.Stack[:len(vm.Stack)]...)
	}
}

func (vm *VM) peek(distance int) value.Value {
	return vm.Stack[len(vm.Stack)-1-distance]
}

func (vm *VM) call(function *value.ObjFunction, argCount int) bool {
	if argCount != function.Arity {
		vm.runtimeError("Expect %d arguments but got %d.", function.Arity, argCount)
		return false
	}

	frame := CallFrame{Function: function, IP: &((function.Chunk.GetCode())[0]), Slots: len(vm.Stack) - argCount - 1}
	vm.Frames = append(vm.Frames, frame)

	return true
}

func (vm *VM) callValue(callee value.Value, argCount int) bool {
	if callee.IsObj() {
		switch callee.ObjType() {
		case objecttype.OBJ_FUNCTION:
			return vm.call(callee.AsFunction(), argCount)

		case objecttype.OBJ_NATIVE:
			native := callee.AsNative()
			result := native.Function(argCount, vm.Stack[len(vm.Stack)-argCount:])

			for i := 0; i < argCount+1; i++ {
				vm.pop()
			}
			vm.push(result)
			return true

		default:
			// Non-callable object type.

		}
	}

	vm.runtimeError("Can only call functions and classes.")
	return false
}

func isFalsey(val value.Value) bool {
	return val.IsNil() || (val.IsBool() && !val.AsBool())
}

func (vm *VM) concatenate() {
	b := vm.pop().AsGoString()
	a := vm.pop().AsGoString()

	vm.push(value.NewObjString(a + b))
}

func (vm *VM) readByte() byte {
	frame := &vm.Frames[len(vm.Frames)-1]

	returnVal := *(frame.IP)
	frame.IP = unsafecode.Increment(frame.IP, 1)

	return returnVal
}

func (vm *VM) readShort() uint16 {
	bytes := make([]byte, 0)
	for i := 0; i < 4; i++ {
		if i == 2 || i == 3 {
			bytes = append(bytes, 0)
			break
		}
		bytes = append(bytes, vm.readByte())
	}
	return binary.BigEndian.Uint16(bytes)
}

func (vm *VM) readLong() uint32 {
	bytes := make([]byte, 0)
	for i := 0; i < 4; i++ {
		if i == 3 {
			bytes = append(bytes, 0)
			break
		}
		bytes = append(bytes, vm.readByte())
	}
	return binary.LittleEndian.Uint32(bytes)
}

func (vm *VM) readConstant() value.Value {
	frame := &vm.Frames[len(vm.Frames)-1]
	return frame.Function.Chunk.GetConstants().Values[vm.readByte()]
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
	frame := &vm.Frames[len(vm.Frames)-1]
	return frame.Function.Chunk.GetConstants().Values[constantAddress]
}

func (vm *VM) binaryOP(valueType valuetype.ValueType, op func(a, b value.Value) interface{}) {
	b := vm.pop()
	a := vm.pop()
	vm.push(value.New(valueType, op(a, b)))
}

func (vm *VM) resetStack() {
	vm.Frames = make([]CallFrame, 0, FRAMES_INITIAL_SIZE)
	vm.Stack = make([]value.Value, 0, STACK_INITIAL_SIZE)
	vm.Globals = make(map[string]value.Value)
}

func (vm *VM) runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Println()

	for i := len(vm.Frames) - 1; i >= 0; i-- {
		frame := &vm.Frames[i]
		function := frame.Function
		// -1 because the IP is sitting on the next instruction to be
		// executed.
		offset := unsafecode.Diff(frame.IP, &((frame.Function.Chunk.GetCode())[0]))
		line := (function.Chunk.GetLines())[offset]
		fmt.Fprintf(os.Stderr, "[line %d] in ", line)
		if function.Name == nil {
			fmt.Fprintf(os.Stderr, "script\n")
		} else {
			fmt.Fprintf(os.Stderr, "%s()\n", function.Name.String)
		}
	}

	vm.resetStack()
}

func (vm *VM) defineNative(name string, function value.NativeFn) {
	vm.push(value.NewObjString(name))
	vm.push(value.NewObjNative(value.NewNative(function)))
	vm.Globals[vm.Stack[0].AsGoString()] = vm.Stack[1]
	vm.pop()
	vm.pop()
}
