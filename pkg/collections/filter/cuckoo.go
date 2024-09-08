package filter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"

	"github.com/ielm/neostd/pkg/collections"
	"github.com/ielm/neostd/pkg/hash"
)

const (
	bucketSize     = 4   // Number of entries per bucket
	maxKicks       = 500 // Maximum number of kicks during insertion
	fingerprintLen = 8   // Length of fingerprint in bits
)

// CuckooFilter is a space-efficient probabilistic data structure for set membership testing.
// It provides fast add, remove, and lookup operations with a controllable false positive rate.
//
// Example:
//
//	cf, _ := NewCuckooFilter(1000, 0.01)
//	cf.Add([]byte("example"))
//	exists := cf.Contains([]byte("example")) // true
type CuckooFilter struct {
	buckets    []uint32 // Array of buckets, each containing 4 fingerprints
	size       uint64   // Number of buckets
	count      uint64   // Number of items in the filter
	loadFactor float64  // Maximum load factor before resizing
	hasher     hash.Hasher[[]byte]
}

// NewCuckooFilter creates a new Cuckoo filter with the given expected number of elements
// and desired false positive rate.
//
// Example:
//
//	cf, err := NewCuckooFilter(1000, 0.01)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewCuckooFilter(expectedElements int, falsePositiveRate float64) (*CuckooFilter, error) {
	hasher, err := hash.NewSipHasher[[]byte]()
	if err != nil {
		return nil, fmt.Errorf("failed to create default hasher: %w", err)
	}
	return NewCuckooFilterWithHasher(expectedElements, falsePositiveRate, hasher)
}

// NewCuckooFilterWithHasher creates a new Cuckoo filter with the given expected number of elements,
// desired false positive rate, and a custom hasher.
//
// Example:
//
//	customHasher := &MyCustomHasher{}
//	cf, err := NewCuckooFilterWithHasher(1000, 0.01, customHasher)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewCuckooFilterWithHasher(expectedElements int, falsePositiveRate float64, hasher hash.Hasher[[]byte]) (*CuckooFilter, error) {
	if expectedElements <= 0 {
		return nil, errors.New("expected elements must be positive")
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, errors.New("false positive rate must be between 0 and 1")
	}

	size := nextPowerOfTwo(uint64(float64(expectedElements) / falsePositiveRate))
	buckets := make([]uint32, size)

	return &CuckooFilter{
		buckets:    buckets,
		size:       size,
		count:      0,
		loadFactor: falsePositiveRate,
		hasher:     hasher,
	}, nil
}

// Add inserts an element into the Cuckoo filter.
// Returns true if the element was successfully inserted, false otherwise.
//
// Example:
//
//	success := cf.Add([]byte("example"))
//	if !success {
//	    fmt.Println("Failed to insert element")
//	}
func (cf *CuckooFilter) Add(data []byte) bool {
	fp := cf.fingerprint(data)
	i1 := cf.index(data)
	i2 := cf.altIndex(i1, fp)

	if cf.insertIntoBucket(i1, fp) || cf.insertIntoBucket(i2, fp) {
		cf.count++
		return true
	}

	// Perform cuckoo hashing
	i := i1
	for k := 0; k < maxKicks; k++ {
		j := uint32(fastrand() % bucketSize)
		fp, cf.buckets[i] = extractFingerprint(cf.buckets[i], j), insertFingerprint(cf.buckets[i], j, fp)
		i = cf.altIndex(i, fp)
		if cf.insertIntoBucket(i, fp) {
			cf.count++
			return true
		}
	}

	return false
}

// Contains checks if an element might be in the Cuckoo filter.
// Note that false positives are possible, but false negatives are not.
//
// Example:
//
//	if cf.Contains([]byte("example")) {
//	    fmt.Println("Element might be in the filter")
//	}
func (cf *CuckooFilter) Contains(data []byte) bool {
	fp := cf.fingerprint(data)
	i1 := cf.index(data)
	i2 := cf.altIndex(i1, fp)
	return cf.containsInBucket(i1, fp) || cf.containsInBucket(i2, fp)
}

// Remove removes an element from the Cuckoo filter.
// Returns true if the element was successfully removed, false if it was not found.
//
// Example:
//
//	removed := cf.Remove([]byte("example"))
//	if removed {
//	    fmt.Println("Element removed from the filter")
//	}
func (cf *CuckooFilter) Remove(data []byte) bool {
	fp := cf.fingerprint(data)
	i1 := cf.index(data)
	i2 := cf.altIndex(i1, fp)

	if cf.removeFromBucket(i1, fp) || cf.removeFromBucket(i2, fp) {
		cf.count--
		return true
	}

	return false
}

// Clear removes all elements from the Cuckoo filter.
//
// Example:
//
//	cf.Clear()
func (cf *CuckooFilter) Clear() {
	for i := range cf.buckets {
		cf.buckets[i] = 0
	}
	cf.count = 0
}

// Size returns the number of items in the filter.
//
// Example:
//
//	count := cf.Size()
//	fmt.Printf("Filter contains %d elements\n", count)
func (cf *CuckooFilter) Size() int {
	return int(cf.count)
}

// IsEmpty returns true if the filter contains no elements.
//
// Example:
//
//	if cf.IsEmpty() {
//	    fmt.Println("Filter is empty")
//	}
func (cf *CuckooFilter) IsEmpty() bool {
	return cf.count == 0
}

// LoadFactor returns the current load factor of the filter.
//
// Example:
//
//	lf := cf.LoadFactor()
//	fmt.Printf("Current load factor: %.2f\n", lf)
func (cf *CuckooFilter) LoadFactor() float64 {
	return float64(cf.count) / (float64(cf.size) * bucketSize)
}

// FalsePositiveRate calculates the current false positive rate of the Cuckoo filter.
//
// Example:
//
//	fpr := cf.FalsePositiveRate()
//	fmt.Printf("Current false positive rate: %.4f\n", fpr)
func (cf *CuckooFilter) FalsePositiveRate() float64 {
	e := 2 * bucketSize * float64(cf.size)
	return 1 - math.Pow(1-1/e, float64(cf.count))
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// It allows the CuckooFilter to be serialized into a binary format.
//
// Example:
//
//	data, err := cf.MarshalBinary()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Save 'data' to a file or send over network
func (cf *CuckooFilter) MarshalBinary() ([]byte, error) {
	data := make([]byte, 24+len(cf.buckets)*4)
	binary.LittleEndian.PutUint64(data[0:8], cf.size)
	binary.LittleEndian.PutUint64(data[8:16], cf.count)
	binary.LittleEndian.PutUint64(data[16:24], math.Float64bits(cf.loadFactor))
	for i, v := range cf.buckets {
		binary.LittleEndian.PutUint32(data[24+i*4:], v)
	}
	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// It allows a CuckooFilter to be deserialized from a binary format.
//
// Example:
//
//	var cf CuckooFilter
//	err := cf.UnmarshalBinary(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (cf *CuckooFilter) UnmarshalBinary(data []byte) error {
	if len(data) < 24 {
		return errors.New("invalid data length")
	}
	cf.size = binary.LittleEndian.Uint64(data[0:8])
	cf.count = binary.LittleEndian.Uint64(data[8:16])
	cf.loadFactor = math.Float64frombits(binary.LittleEndian.Uint64(data[16:24]))
	cf.buckets = make([]uint32, (len(data)-24)/4)
	for i := range cf.buckets {
		cf.buckets[i] = binary.LittleEndian.Uint32(data[24+i*4:])
	}
	var err error
	cf.hasher, err = hash.NewSipHasher[[]byte]()
	return err
}

// Helper functions

func (cf *CuckooFilter) fingerprint(data []byte) uint8 {
	h, _ := cf.hasher.Hash(data)
	return uint8(h&0xFF) | 1 // Ensure fingerprint is non-zero
}

func (cf *CuckooFilter) index(data []byte) uint64 {
	h, _ := cf.hasher.Hash(data)
	return h % cf.size
}

func (cf *CuckooFilter) altIndex(i uint64, fp uint8) uint64 {
	h := uint64(fp) * 0x5bd1e995 // MurmurHash2 constant
	return (i ^ h) % cf.size
}

func (cf *CuckooFilter) insertIntoBucket(i uint64, fp uint8) bool {
	if emptySlot := bits.TrailingZeros32(^cf.buckets[i]) / fingerprintLen; emptySlot < bucketSize {
		cf.buckets[i] = insertFingerprint(cf.buckets[i], uint32(emptySlot), fp)
		return true
	}
	return false
}

func (cf *CuckooFilter) containsInBucket(i uint64, fp uint8) bool {
	return (cf.buckets[i]&uint32(fp)) != 0 && containsFingerprint(cf.buckets[i], fp)
}

func (cf *CuckooFilter) removeFromBucket(i uint64, fp uint8) bool {
	if slot := findFingerprint(cf.buckets[i], fp); slot < bucketSize {
		cf.buckets[i] = removeFingerprint(cf.buckets[i], uint32(slot))
		return true
	}
	return false
}

func insertFingerprint(bucket uint32, slot uint32, fp uint8) uint32 {
	return bucket | (uint32(fp) << (slot * fingerprintLen))
}

func extractFingerprint(bucket uint32, slot uint32) uint8 {
	return uint8((bucket >> (slot * fingerprintLen)) & 0xFF)
}

func removeFingerprint(bucket uint32, slot uint32) uint32 {
	return bucket & ^(0xFF << (slot * fingerprintLen))
}

func containsFingerprint(bucket uint32, fp uint8) bool {
	return (bucket&uint32(fp)) != 0 && findFingerprint(bucket, fp) < bucketSize
}

func findFingerprint(bucket uint32, fp uint8) uint32 {
	x := bucket ^ (uint32(fp) * 0x01010101)
	return uint32(bits.OnesCount32((x - 0x01010101) & ^x & 0x80808080))
}

// Ensure CuckooFilter implements the ProbabilisticSet interface
var _ collections.ProbabilisticSet[[]byte] = (*CuckooFilter)(nil)
