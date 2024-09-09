package hash

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

// SipHasher implements the SipHash 1-3 algorithm
type SipHasher struct {
	BaseHasher
	k0, k1 uint64
}

// NewSipHasher creates a new SipHasher with random keys
func NewSipHasher() (*SipHasher, error) {
	k0, k1, err := GenerateRandomKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random keys: %w", err)
	}
	return &SipHasher{k0: k0, k1: k1}, nil
}

// Write adds more data to the running hash
func (s *SipHasher) Write(p []byte) (n int, err error) {
	s.BaseHasher.Write(p)
	return len(p), nil
}

// Sum appends the current hash to b and returns the resulting slice
func (s *SipHasher) Sum(b []byte) []byte {
	h := s.sipHash13(s.BaseHasher.Sum(nil))
	return append(b, Uint64ToBytes(h)...)
}

// Reset resets the hash to its initial state
func (s *SipHasher) Reset() {
	s.BaseHasher.Reset()
}

// Size returns the number of bytes Sum will return
func (s *SipHasher) Size() int {
	return 8
}

// BlockSize returns the hash's underlying block size
func (s *SipHasher) BlockSize() int {
	return 64
}

// sipHash13 implements the core SipHash-1-3 algorithm
func (s *SipHasher) sipHash13(data []byte) uint64 {
	v0, v1, v2, v3 := s.initializeState()
	dataLen := uint64(len(data))

	// Process full 64-bit blocks
	for len(data) >= 8 {
		m := binary.LittleEndian.Uint64(data)
		v3 ^= m
		s.sipRound(&v0, &v1, &v2, &v3)
		v0 ^= m
		data = data[8:]
	}

	// Process remaining bytes and finalize
	v3 ^= s.encodeLastBlock(data, dataLen)
	v0, v1, v2, v3 = s.finalize(v0, v1, v2, v3)

	return v0 ^ v1 ^ v2 ^ v3
}

// initializeState sets up the initial state for SipHash
func (s *SipHasher) initializeState() (uint64, uint64, uint64, uint64) {
	return s.k0 ^ 0x736f6d6570736575,
		s.k1 ^ 0x646f72616e646f6d,
		s.k0 ^ 0x6c7967656e657261,
		s.k1 ^ 0x7465646279746573
}

// encodeLastBlock processes the remaining bytes and encodes the data length
func (s *SipHasher) encodeLastBlock(data []byte, dataLen uint64) uint64 {
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
func (s *SipHasher) finalize(v0, v1, v2, v3 uint64) (uint64, uint64, uint64, uint64) {
	v2 ^= 0xff
	for i := 0; i < 3; i++ {
		s.sipRound(&v0, &v1, &v2, &v3)
	}
	return v0, v1, v2, v3
}

// sipRound performs a single round of the SipHash algorithm
func (s *SipHasher) sipRound(v0, v1, v2, v3 *uint64) {
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
