package unsafecode

import (
	"golanglox/lib/value"
	"unsafe"
)

func IndexSlot(pointer *value.Value, index int) *value.Value {
	// Convert a pointer to an byte to an unsafe.Pointer, then to a uintptr.
	addressHolder := uintptr(unsafe.Pointer(pointer))

	// Increment the value of the address by the number of bytes of an element
	addressHolder = addressHolder + unsafe.Sizeof(*(pointer))*uintptr(index)

	// Convert a uintptr to an unsafe.Pointer, then to a pointer to an byte.
	return (*value.Value)(unsafe.Pointer(addressHolder))
}

func Increment(pointer *byte, steps int) *byte {
	// Convert a pointer to an byte to an unsafe.Pointer, then to a uintptr.
	addressHolder := uintptr(unsafe.Pointer(pointer))

	// Increment the value of the address by the number of bytes of an element
	addressHolder = addressHolder + unsafe.Sizeof(*(pointer))*uintptr(steps)

	// Convert a uintptr to an unsafe.Pointer, then to a pointer to an byte.
	return (*byte)(unsafe.Pointer(addressHolder))
}

func Decrement(pointer *byte, steps int) *byte {
	// Convert a pointer to an byte to an unsafe.Pointer, then to a uintptr.
	addressHolder := uintptr(unsafe.Pointer(pointer))

	// Increment the value of the address by the number of bytes of an element
	addressHolder = addressHolder - unsafe.Sizeof(*(pointer))*uintptr(steps)

	// Convert a uintptr to an unsafe.Pointer, then to a pointer to an byte.
	return (*byte)(unsafe.Pointer(addressHolder))
}

func Diff(p1 *byte, p2 *byte) int {
	return int(uintptr(unsafe.Pointer(p1)) - uintptr(unsafe.Pointer(p2)))
}
