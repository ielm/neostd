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

// BloomFilter is a space-efficient probabilistic data structure that provides
// efficient membership testing on a set.
type BloomFilter struct {
	bitset    []uint64
	size      uint64
	hashCount uint64
	hasher    hash.Hasher[[]byte]
}

// NewBloomFilter creates a new Bloom filter with the given expected number of elements
// and desired false positive rate.
//
// Example:
//
//	bf, err := NewBloomFilter(1000, 0.01)
//	if err != nil {
//		log.Fatal(err)
//	}
func NewBloomFilter(expectedElements int, falsePositiveRate float64) (*BloomFilter, error) {
	hasher, err := hash.NewSipHasher[[]byte]()
	if err != nil {
		return nil, fmt.Errorf("failed to create default hasher: %w", err)
	}
	return NewBloomFilterWithHasher(expectedElements, falsePositiveRate, hasher)
}

// NewBloomFilterWithHasher creates a new Bloom filter with the given expected number of elements,
// desired false positive rate, and a custom hasher.
//
// Example:
//
//	customHasher := &MyCustomHasher{}
//	bf, err := NewBloomFilterWithHasher(1000, 0.01, customHasher)
//	if err != nil {
//		log.Fatal(err)
//	}
func NewBloomFilterWithHasher(expectedElements int, falsePositiveRate float64, hasher hash.Hasher[[]byte]) (*BloomFilter, error) {
	if expectedElements <= 0 {
		return nil, errors.New("expected elements must be positive")
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, errors.New("false positive rate must be between 0 and 1")
	}

	size := optimalSize(expectedElements, falsePositiveRate)
	hashCount := optimalHashCount(size, expectedElements)

	return &BloomFilter{
		bitset:    make([]uint64, (size+63)/64),
		size:      size,
		hashCount: hashCount,
		hasher:    hasher,
	}, nil
}

// Add inserts an element into the Bloom filter.
// It returns true if the element was not present before, false otherwise.
//
// Example:
//
//	wasNew := bf.Add([]byte("example"))
func (bf *BloomFilter) Add(data []byte) bool {
	h1, h2 := bf.hashValues(data)
	allSet := true
	for i := uint64(0); i < bf.hashCount; i++ {
		index := bf.index(h1, h2, i)
		if !bf.getBit(index) {
			allSet = false
			bf.setBit(index)
		}
	}
	return !allSet
}

// Contains checks if an element might be in the Bloom filter.
//
// Example:
//
//	if bf.Contains([]byte("example")) {
//		fmt.Println("Element might be in the set")
//	}
func (bf *BloomFilter) Contains(data []byte) bool {
	h1, h2 := bf.hashValues(data)
	for i := uint64(0); i < bf.hashCount; i++ {
		if !bf.getBit(bf.index(h1, h2, i)) {
			return false
		}
	}
	return true
}

// Clear removes all elements from the Bloom filter.
//
// Example:
//
//	bf.Clear()
func (bf *BloomFilter) Clear() {
	for i := range bf.bitset {
		bf.bitset[i] = 0
	}
}

// EstimateElementCount estimates the number of elements in the Bloom filter.
//
// Example:
//
//	count := bf.EstimateElementCount()
//	fmt.Printf("Estimated number of elements: %d\n", count)
func (bf *BloomFilter) EstimateElementCount() uint64 {
	setBits := bf.countSetBits()
	return uint64(-(float64(bf.size) / float64(bf.hashCount)) * math.Log(1-float64(setBits)/float64(bf.size)))
}

// FalsePositiveRate calculates the current false positive rate of the Bloom filter.
//
// Example:
//
//	fpr := bf.FalsePositiveRate()
//	fmt.Printf("Current false positive rate: %.4f\n", fpr)
func (bf *BloomFilter) FalsePositiveRate() float64 {
	setBits := float64(bf.countSetBits())
	return math.Pow(setBits/float64(bf.size), float64(bf.hashCount))
}

func (bf *BloomFilter) hashValues(data []byte) (uint64, uint64) {
	hashBytes, err := bf.hasher.Hash(data)
	if err != nil {
		panic(err) // In production, consider a more graceful error handling
	}
	h1 := binary.LittleEndian.Uint64(hashBytes)
	h2 := h1 >> 32
	return h1, h2
}

// index calculates the bit index for the i-th hash function.
func (bf *BloomFilter) index(h1, h2, i uint64) uint64 {
	return (h1 + i*h2) % bf.size
}

// setBit sets the bit at the given index.
func (bf *BloomFilter) setBit(index uint64) {
	bf.bitset[index/64] |= 1 << (index % 64)
}

// getBit checks if the bit at the given index is set.
func (bf *BloomFilter) getBit(index uint64) bool {
	return bf.bitset[index/64]&(1<<(index%64)) != 0
}

// countSetBits counts the number of set bits in the Bloom filter.
func (bf *BloomFilter) countSetBits() uint64 {
	var count uint64
	for _, x := range bf.bitset {
		count += uint64(bits.OnesCount64(x))
	}
	return count
}

// optimalSize calculates the optimal size of the Bloom filter.
func optimalSize(n int, p float64) uint64 {
	return uint64(math.Ceil(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

// optimalHashCount calculates the optimal number of hash functions.
func optimalHashCount(size uint64, n int) uint64 {
	return uint64(math.Ceil(float64(size) / float64(n) * math.Log(2)))
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// It serializes the Bloom filter into a binary format.
//
// Example:
//
//	data, err := bf.MarshalBinary()
//	if err != nil {
//		log.Fatal(err)
//	}
func (bf *BloomFilter) MarshalBinary() ([]byte, error) {
	data := make([]byte, 16+len(bf.bitset)*8)
	binary.LittleEndian.PutUint64(data[0:8], bf.size)
	binary.LittleEndian.PutUint64(data[8:16], bf.hashCount)
	for i, v := range bf.bitset {
		binary.LittleEndian.PutUint64(data[16+i*8:], v)
	}
	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// It deserializes the Bloom filter from a binary format.
//
// Example:
//
//	err := bf.UnmarshalBinary(data)
//	if err != nil {
//		log.Fatal(err)
//	}
func (bf *BloomFilter) UnmarshalBinary(data []byte) error {
	if len(data) < 16 {
		return errors.New("invalid data length")
	}
	bf.size = binary.LittleEndian.Uint64(data[0:8])
	bf.hashCount = binary.LittleEndian.Uint64(data[8:16])
	bf.bitset = make([]uint64, (bf.size+63)/64)
	for i := range bf.bitset {
		bf.bitset[i] = binary.LittleEndian.Uint64(data[16+i*8:])
	}
	var err error
	bf.hasher, err = hash.NewSipHasher[[]byte]()
	return err
}

// Size returns the current number of elements in the Bloom filter.
//
// Example:
//
//	size := bf.Size()
//	fmt.Printf("Number of elements: %d\n", size)
func (bf *BloomFilter) Size() int {
	return int(bf.EstimateElementCount())
}

// IsEmpty returns true if the Bloom filter contains no elements.
//
// Example:
//
//	if bf.IsEmpty() {
//		fmt.Println("Bloom filter is empty")
//	}
func (bf *BloomFilter) IsEmpty() bool {
	return bf.EstimateElementCount() == 0
}

// Capacity returns the maximum number of elements the Bloom filter can hold
// while maintaining the desired false positive rate.
//
// Example:
//
//	capacity := bf.Capacity()
//	fmt.Printf("Maximum capacity: %d\n", capacity)
func (bf *BloomFilter) Capacity() int {
	return int(float64(bf.size) * math.Log(2) / float64(bf.hashCount))
}

// Merge combines this Bloom filter with another one of the same size and hash count.
//
// Example:
//
//	err := bf1.Merge(bf2)
//	if err != nil {
//		log.Fatal(err)
//	}
func (bf *BloomFilter) Merge(other *BloomFilter) error {
	if bf.size != other.size || bf.hashCount != other.hashCount {
		return errors.New("bloom filters must have the same size and hash count to merge")
	}
	for i := range bf.bitset {
		bf.bitset[i] |= other.bitset[i]
	}
	return nil
}

// Copy creates a deep copy of the Bloom filter.
//
// Example:
//
//	newBF := bf.Copy()
func (bf *BloomFilter) Copy() *BloomFilter {
	newBF := &BloomFilter{
		bitset:    make([]uint64, len(bf.bitset)),
		size:      bf.size,
		hashCount: bf.hashCount,
		hasher:    bf.hasher,
	}
	copy(newBF.bitset, bf.bitset)
	return newBF
}

// Ensure BloomFilter implements the ProbabilisticSet interface
var _ collections.ProbabilisticSet[[]byte] = (*BloomFilter)(nil)
