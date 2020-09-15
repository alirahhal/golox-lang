package unsafecode

import (
	"unsafe"
)

func Increment(pointer *byte) *byte {
	// Convert a pointer to an byte to an unsafe.Pointer, then to a uintptr.
	addressHolder := uintptr(unsafe.Pointer(pointer))

	// Increment the value of the address by the number of bytes of an element
	addressHolder = addressHolder + unsafe.Sizeof(*(pointer))

	// Convert a uintptr to an unsafe.Pointer, then to a pointer to an byte.
	return (*byte)(unsafe.Pointer(addressHolder))
}
