package hash

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"unsafe"
)

// Hasher is an interface that extends the standard hash.Hash interface
type Hasher interface {
	hash.Hash
	HashKey(key any) ([]byte, error)
}

// BaseHasher is a struct that implements the Hasher interface
type BaseHasher struct {
	hash.Hash
}

// HashKey converts a key of any type to a byte slice and then hashes it
func (bh *BaseHasher) HashKey(key any) ([]byte, error) {
	data, err := keyToBytes(key)
	if err != nil {
		return nil, err
	}
	bh.Reset()
	_, err = bh.Write(data)
	if err != nil {
		return nil, err
	}
	return bh.Sum(nil), nil
}

// HashBytesToUint64 converts a byte slice to uint64
func HashBytesToUint64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

// keyToBytes converts a key of any type to a byte slice
func keyToBytes(key any) ([]byte, error) {
	switch k := key.(type) {
	case string:
		return []byte(k), nil
	case []byte:
		return k, nil
	default:
		return ToBinary(k)
	}
}

// ToBinary converts an interface{} to a byte slice
func ToBinary(v interface{}) ([]byte, error) {
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

// GenerateRandomKeys creates two cryptographically secure random uint64 values
func GenerateRandomKeys() (uint64, uint64, error) {
	var b [16]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read random bytes: %w", err)
	}
	return binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:]), nil
}

// Uint64ToBytes converts a uint64 to a byte slice
func Uint64ToBytes(value uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, value)
	return b
}
