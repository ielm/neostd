package hash

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/bits"
	"unsafe"
)

// Hasher is an interface for hash functions
type Hasher[K any] interface {
	Hash(key K) ([]byte, error)
}

// SipHasher implements the SipHash 1-3 algorithm
type SipHasher[K any] struct {
	k0, k1 uint64
}

// NewSipHasher creates a new SipHasher with random keys
func NewSipHasher[K any]() (*SipHasher[K], error) {
	k0, k1, err := generateRandomKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random keys: %w", err)
	}
	return &SipHasher[K]{k0: k0, k1: k1}, nil
}

// Hash computes the hash of the given key using SipHash 1-3
func (s *SipHasher[K]) Hash(key K) ([]byte, error) {
	data, err := keyToBytes(key)
	if err != nil {
		return nil, fmt.Errorf("failed to convert key to bytes: %w", err)
	}
	hashValue := sipHash13(s.k0, s.k1, data)
	return uint64ToBytes(hashValue), nil
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

// sipHash13 implements the core SipHash-1-3 algorithm
func sipHash13(k0, k1 uint64, data []byte) uint64 {
	v0, v1, v2, v3 := initializeState(k0, k1)
	dataLen := uint64(len(data))

	// Process full 64-bit blocks
	for len(data) >= 8 {
		m := binary.LittleEndian.Uint64(data)
		v3 ^= m
		sipRound(&v0, &v1, &v2, &v3)
		v0 ^= m
		data = data[8:]
	}

	// Process remaining bytes and finalize
	v3 ^= encodeLastBlock(data, dataLen)
	v0, v1, v2, v3 = finalize(v0, v1, v2, v3)

	return v0 ^ v1 ^ v2 ^ v3
}

// initializeState sets up the initial state for SipHash
func initializeState(k0, k1 uint64) (uint64, uint64, uint64, uint64) {
	return k0 ^ 0x736f6d6570736575,
		k1 ^ 0x646f72616e646f6d,
		k0 ^ 0x6c7967656e657261,
		k1 ^ 0x7465646279746573
}

// encodeLastBlock processes the remaining bytes and encodes the data length
func encodeLastBlock(data []byte, dataLen uint64) uint64 {
	var t uint64
	switch len(data) {
	case 7:
		t |= uint64(data[6]) << 48
		fallthrough
	case 6:
		t |= uint64(data[5]) << 40
		fallthrough
	case 5:
		t |= uint64(data[4]) << 32
		fallthrough
	case 4:
		t |= uint64(data[3]) << 24
		fallthrough
	case 3:
		t |= uint64(data[2]) << 16
		fallthrough
	case 2:
		t |= uint64(data[1]) << 8
		fallthrough
	case 1:
		t |= uint64(data[0])
	}
	return t | (dataLen << 56)
}

// finalize performs the final rounds of SipHash
func finalize(v0, v1, v2, v3 uint64) (uint64, uint64, uint64, uint64) {
	v2 ^= 0xff
	for i := 0; i < 3; i++ {
		sipRound(&v0, &v1, &v2, &v3)
	}
	return v0, v1, v2, v3
}

// sipRound performs a single round of the SipHash algorithm
func sipRound(v0, v1, v2, v3 *uint64) {
	*v0 += *v1
	*v1 = bits.RotateLeft64(*v1, 13)
	*v1 ^= *v0
	*v0 = bits.RotateLeft64(*v0, 32)
	*v2 += *v3
	*v3 = bits.RotateLeft64(*v3, 16)
	*v3 ^= *v2
	*v0 += *v3
	*v3 = bits.RotateLeft64(*v3, 21)
	*v3 ^= *v0
	*v2 += *v1
	*v1 = bits.RotateLeft64(*v1, 17)
	*v1 ^= *v2
	*v2 = bits.RotateLeft64(*v2, 32)
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
func generateRandomKeys() (uint64, uint64, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read random bytes: %w", err)
	}
	return binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:]), nil
}

// uint64ToBytes converts a uint64 to a byte slice
func uint64ToBytes(value uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, value)
	return b
}
