package hash

import (
	"hash"
)

// Hasher is an interface that defines the methods required for a hash function
type Hasher interface {
	Hash(data []byte) uint64
	New() hash.Hash64
}

// DefaultHasher is the default implementation using SipHash
type DefaultHasher struct {
	key [16]byte
}

// NewDefaultHasher creates a new DefaultHasher with the given key
func NewDefaultHasher(key [16]byte) *DefaultHasher {
	return &DefaultHasher{key: key}
}

// Hash implements the Hasher interface for DefaultHasher
func (h *DefaultHasher) Hash(data []byte) uint64 {
	return Sum64(data, &h.key)
}

// New implements the Hasher interface for DefaultHasher
func (h *DefaultHasher) New() hash.Hash64 {
	hasher, _ := New64(h.key[:])
	return hasher
}
