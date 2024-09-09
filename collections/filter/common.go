package filter

import (
	"math/bits"
	"unsafe"
)

// nextPowerOfTwo calculates the next power of two for a given number.
func nextPowerOfTwo(x uint64) uint64 {
	return 1 << (64 - bits.LeadingZeros64(x-1))
}

// fastrand is a fast, thread-safe random number generator.
func fastrand() uint32 {
	return uint32(uintptr(unsafe.Pointer(new(byte))))*1664525 + 1013904223
}
