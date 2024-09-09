package hash

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Hasher is an interface for hash functions
type Hasher[K any] interface {
	Hash(key K) ([]byte, error)
}

// hash bytes to uint64
func HashBytesToUint64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

// keyToBytes converts a key of any type to a byte slice
func keyToBytes[K any](key K) ([]byte, error) {
	switch k := any(key).(type) {
	case string:
		return []byte(k), nil
	case []byte:
		return k, nil
	default:
		return toBinary(k)
	}
}

// toBinary converts an interface{} to a byte slice
func toBinary(v interface{}) ([]byte, error) {
	size := int(unsafe.Sizeof(v))
	if size > 1<<30 {
		return nil, fmt.Errorf("input size too large: %d bytes", size)
	}
	b := make([]byte, size)
	switch size {
	case 1:
		b[0] = *(*uint8)(unsafe.Pointer(&v))
	case 2:
		binary.LittleEndian.PutUint16(b, *(*uint16)(unsafe.Pointer(&v)))
	case 4:
		binary.LittleEndian.PutUint32(b, *(*uint32)(unsafe.Pointer(&v)))
	case 8:
		binary.LittleEndian.PutUint64(b, *(*uint64)(unsafe.Pointer(&v)))
	default:
		copy(b, (*[1 << 30]byte)(unsafe.Pointer(&v))[:size])
	}
	return b, nil
}

// generateRandomKeys creates two cryptographically secure random uint64 values
func GenerateRandomKeys() (uint64, uint64, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read random bytes: %w", err)
	}
	return binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:]), nil
}

// uint64ToBytes converts a uint64 to a byte slice
func Uint64ToBytes(value uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, value)
	return b
}
