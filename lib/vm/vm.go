package vm

import (
	"fmt"
	"golanglox/lib/chunk"
	"golanglox/lib/chunk/opcode"
	"golanglox/lib/config"
	"golanglox/lib/debug"
	"golanglox/lib/value"
	"golanglox/lib/vm/interpretresult"
	"unsafe"
)

type VM struct {
	Chunk *chunk.Chunk
	IP    *byte
}

func InitVM() *VM {
	return &VM{Chunk: nil}
}

func (vm *VM) Interpret(chunk *chunk.Chunk) interpretresult.InterpretResult {
	vm.Chunk = chunk
	vm.IP = &(vm.Chunk.Code[0])
	return vm.run()
}

func (vm *VM) FreeVM() {}

func (vm *VM) run() interpretresult.InterpretResult {
	for {
		if config.DEBUG_TRACE_EXECUTION {
			debug.DisassembleInstruction(vm.Chunk, int(uintptr(unsafe.Pointer(vm.IP))-uintptr(unsafe.Pointer(&(vm.Chunk.Code[0])))))
		}
		var instruction byte
		switch instruction = vm.readByte(); instruction {
		case opcode.OP_RETURN:
			return interpretresult.INTERPRET_OK
		case opcode.OP_CONSTANT:
			var constant value.Value = vm.readConstant()
			constant.PrintValue()
			fmt.Print("\n")
			break
		}
	}
}

func (vm *VM) readByte() byte {
	// Convert a pointer to an byte to an unsafe.Pointer, then to a uintptr.
	addressHolder := uintptr(unsafe.Pointer(vm.IP))

	// Increment the value of the address by the number of bytes of an element which is an byte.
	addressHolder = addressHolder + unsafe.Sizeof(*(vm.IP))

	// Convert a uintptr to an unsafe.Pointer, then to a pointer to an byte.
	newPtr := (*byte)(unsafe.Pointer(addressHolder))

	returnVal := *(vm.IP)
	vm.IP = newPtr

	return returnVal
}

func (vm *VM) readConstant() value.Value {
	return vm.Chunk.Constants.Values[vm.readByte()]
}
