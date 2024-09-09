package filter

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/hash"
)

const (
	segmentLength = 256
	blockLength   = 64
)

// XorFilter is a space-efficient probabilistic data structure for set membership testing.
// It provides fast, constant-time operations for adding elements and testing membership,
// with a controllable false positive rate.
type XorFilter struct {
	fingerprints       []uint8
	blockLength        uint32
	segmentLength      uint32
	segmentLengthMask  uint32
	segmentCount       uint32
	segmentCountLength uint32
	seed               uint64
	hasher             hash.Hasher
}

// NewXorFilter creates a new Xor filter with the given expected number of elements.
// It initializes the filter with an optimal size based on the expected number of elements.
//
// Example:
//
//	xf, err := NewXorFilter(1000000) // Create a filter expecting 1 million elements
//	if err != nil {
//		log.Fatal(err)
//	}
func NewXorFilter(expectedElements int) (*XorFilter, error) {
	hasher, err := hash.NewSipHasher()
	if err != nil {
		return nil, err
	}
	return NewXorFilterWithHasher(expectedElements, hasher)
}

// NewXorFilterWithHasher creates a new Xor filter with the given expected number of elements
// and a custom hasher.
//
// Example:
//
//	customHasher := &MyCustomHasher{}
//	xf, err := NewXorFilterWithHasher(1000000, customHasher)
//	if err != nil {
//		log.Fatal(err)
//	}
func NewXorFilterWithHasher(expectedElements int, hasher hash.Hasher) (*XorFilter, error) {
	if expectedElements <= 0 {
		return nil, errors.New("expected elements must be positive")
	}

	capacity := nextPowerOfTwo(uint64(math.Ceil(float64(expectedElements) * 1.23)))
	segmentCount := capacity / segmentLength

	return &XorFilter{
		fingerprints:       make([]uint8, capacity),
		blockLength:        blockLength,
		segmentLength:      segmentLength,
		segmentLengthMask:  segmentLength - 1,
		segmentCount:       uint32(segmentCount),
		segmentCountLength: uint32(segmentCount * segmentLength),
		seed:               0,
		hasher:             hasher,
	}, nil
}

// Add inserts an element into the Xor filter.
// Note: Xor filters don't support dynamic insertion after construction.
// This method is a no-op to satisfy the ProbabilisticSet interface.
//
// Example:
//
//	added := xf.Add([]byte("example"))
//	// added will always be false for XorFilter
func (xf *XorFilter) Add(data []byte) bool {
	// Xor filters don't support dynamic insertion.
	// This method is a no-op to satisfy the ProbabilisticSet interface.
	return false
}

// Contains checks if an element might be in the Xor filter.
// It may return false positives, but never false negatives.
//
// Example:
//
//	if xf.Contains([]byte("example")) {
//		fmt.Println("Element might be in the set")
//	}
func (xf *XorFilter) Contains(data []byte) bool {
	h1, h2, h3 := xf.hashValues(data)
	f := xf.fingerprint(h1)
	return xf.fingerprints[h1]^xf.fingerprints[h2]^xf.fingerprints[h3] == f
}

// Clear removes all elements from the Xor filter.
//
// Example:
//
//	xf.Clear()
func (xf *XorFilter) Clear() {
	for i := range xf.fingerprints {
		xf.fingerprints[i] = 0
	}
}

// Size returns the number of items in the filter.
//
// Example:
//
//	count := xf.Size()
func (xf *XorFilter) Size() int {
	return int(xf.segmentCountLength)
}

// IsEmpty returns true if the filter contains no elements.
//
// Example:
//
//	if xf.IsEmpty() {
//		fmt.Println("XorFilter is empty")
//	}
func (xf *XorFilter) IsEmpty() bool {
	for _, fp := range xf.fingerprints {
		if fp != 0 {
			return false
		}
	}
	return true
}

// FalsePositiveRate calculates the current false positive rate of the Xor filter.
//
// Example:
//
//	fpr := xf.FalsePositiveRate()
//	fmt.Printf("False positive rate: %.6f\n", fpr)
func (xf *XorFilter) FalsePositiveRate() float64 {
	return 1.0 / float64(1<<8) // 1/256 for 8-bit fingerprints
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// It serializes the XorFilter into a binary format.
//
// Example:
//
//	data, err := xf.MarshalBinary()
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Use 'data' for storage or transmission
func (xf *XorFilter) MarshalBinary() ([]byte, error) {
	data := make([]byte, 28+len(xf.fingerprints))
	binary.LittleEndian.PutUint32(data[0:4], xf.blockLength)
	binary.LittleEndian.PutUint32(data[4:8], xf.segmentLength)
	binary.LittleEndian.PutUint32(data[8:12], xf.segmentLengthMask)
	binary.LittleEndian.PutUint32(data[12:16], xf.segmentCount)
	binary.LittleEndian.PutUint32(data[16:20], xf.segmentCountLength)
	binary.LittleEndian.PutUint64(data[20:28], xf.seed)
	copy(data[28:], xf.fingerprints)
	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// It deserializes the XorFilter from a binary format.
//
// Example:
//
//	var newXF XorFilter
//	err := newXF.UnmarshalBinary(data)
//	if err != nil {
//		log.Fatal(err)
//	}
func (xf *XorFilter) UnmarshalBinary(data []byte) error {
	if len(data) < 28 {
		return errors.New("invalid data length")
	}
	xf.blockLength = binary.LittleEndian.Uint32(data[0:4])
	xf.segmentLength = binary.LittleEndian.Uint32(data[4:8])
	xf.segmentLengthMask = binary.LittleEndian.Uint32(data[8:12])
	xf.segmentCount = binary.LittleEndian.Uint32(data[12:16])
	xf.segmentCountLength = binary.LittleEndian.Uint32(data[16:20])
	xf.seed = binary.LittleEndian.Uint64(data[20:28])
	xf.fingerprints = make([]uint8, len(data)-28)
	copy(xf.fingerprints, data[28:])

	// Use the default SipHasher for deserialized XorFilters
	var err error
	xf.hasher, err = hash.NewSipHasher()
	return err
}

// Helper functions

func (xf *XorFilter) hashValues(data []byte) (uint32, uint32, uint32) {
	xf.hasher.Reset()
	xf.hasher.Write(data)
	h := hash.HashBytesToUint64(xf.hasher.Sum(nil))
	h1 := uint32(h) & (xf.segmentCountLength - 1)
	h2 := uint32(h>>32) & (xf.segmentCountLength - 1)
	h3 := xf.hash(uint64(h1) ^ uint64(h2))
	return h1, h2, h3
}

func (xf *XorFilter) hash(x uint64) uint32 {
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x = x ^ (x >> 31)
	return uint32(x) & (xf.segmentCountLength - 1)
}

func (xf *XorFilter) fingerprint(hash uint32) uint8 {
	return uint8(hash>>7 | 1)
}

// Ensure XorFilter implements the ProbabilisticSet interface
var _ collections.ProbabilisticSet[[]byte] = (*XorFilter)(nil)
