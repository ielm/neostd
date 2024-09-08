package maps

import (
	"math/bits"
	"reflect"
	"unsafe"

	"github.com/ielm/neostd/pkg/collections"
	"github.com/ielm/neostd/pkg/hash"
)

const (
	defaultLoadFactor = 0.875
	minCapacity       = 8
	groupSize         = 16
	maxProbeDistance  = 128
	emptyByte         = 0b11111111
)

// HashMap is a high-performance hash table implementation.
// It uses open addressing with quadratic probing and SIMD-like optimizations.
type HashMap[K any, V any] struct {
	ctrl       []byte
	entries    []entry[K, V]
	size       int
	capacity   int
	loadFactor float64
	hasher     hash.Hasher[K]
	comparator collections.Comparator[K]
}

type entry[K any, V any] struct {
	key   K
	value V
}

// NewHashMap creates a new HashMap with default settings.
func NewHashMap[K any, V any]() *HashMap[K, V] {
	h := &HashMap[K, V]{
		capacity:   minCapacity,
		loadFactor: defaultLoadFactor,
	}
	hasher, err := hash.NewSipHasher[K]()
	if err != nil {
		panic(err) // In production, consider handling this error more gracefully
	}
	h.hasher = hasher
	h.initializeCtrl()
	return h
}

// initializeCtrl initializes the control bytes and entries.
func (h *HashMap[K, V]) initializeCtrl() {
	h.ctrl = make([]byte, h.capacity+groupSize)
	for i := range h.ctrl {
		h.ctrl[i] = emptyByte
	}
	h.entries = make([]entry[K, V], h.capacity)
}

// Put inserts a key-value pair into the HashMap.
// It returns the old value and a boolean indicating if the key existed.
func (h *HashMap[K, V]) Put(key K, value V) (V, bool) {
	if h.shouldResize() {
		h.resize(h.capacity * 2)
	}

	hash := h.hashKey(key)
	index, existed := h.findOrInsert(hash, key)

	oldValue := h.entries[index].value
	h.entries[index] = entry[K, V]{key: key, value: value}

	if !existed {
		h.size++
	}

	return oldValue, existed
}

// shouldResize checks if the HashMap needs to be resized
func (h *HashMap[K, V]) shouldResize() bool {
	return h.size >= int(float64(h.capacity)*h.loadFactor)
}

// findOrInsert finds an existing entry or inserts a new one using quadratic probing.
func (h *HashMap[K, V]) findOrInsert(hash uint64, key K) (int, bool) {
	index := hash & uint64(h.capacity-1)
	hashByte := h.hashToByte(hash)

	for i := uint64(0); i < maxProbeDistance; i++ {
		group := index & ^uint64(groupSize-1)
		match := h.matchGroup(group, hashByte)

		for match != 0 {
			matchIndex := group + uint64(bits.TrailingZeros64(uint64(match)))
			if h.compareKeys(h.entries[matchIndex].key, key) {
				return int(matchIndex), true
			}
			match &= match - 1
		}

		if emptySlot := h.findEmptySlot(group); emptySlot != -1 {
			slotIndex := int(group) + emptySlot
			h.ctrl[slotIndex] = hashByte
			return slotIndex, false
		}

		index = h.nextProbe(index, i)
	}

	// If we reach here, we need to resize and try again
	h.resize(h.capacity * 2)
	return h.findOrInsert(hash, key)
}

// matchGroup performs SIMD-like matching of control bytes.
func (h *HashMap[K, V]) matchGroup(group uint64, hashByte byte) uint16 {
	vec := (*[16]uint8)(unsafe.Pointer(&h.ctrl[group]))
	mask := uint16(0)

	// Perform 16 comparisons in parallel
	// Please no one ask me how this works, I just know it's crazy fast
	for i := 0; i < 16; i += 8 {
		chunk := *(*uint64)(unsafe.Pointer(&vec[i]))
		eq := chunk ^ (uint64(hashByte) * 0x0101010101010101)
		bitmask := ((eq - 0x0101010101010101) & ^eq & 0x8080808080808080) >> 7
		mask |= uint16(bitmask) << i
	}

	return mask
}

// findEmptySlot finds an empty slot in a group.
func (h *HashMap[K, V]) findEmptySlot(group uint64) int {
	vec := (*[16]uint8)(unsafe.Pointer(&h.ctrl[group]))

	// Check 16 slots in parallel
	for i := 0; i < 16; i += 8 {
		chunk := *(*uint64)(unsafe.Pointer(&vec[i]))
		eq := chunk ^ (uint64(emptyByte) * 0x0101010101010101)
		bitmask := ((eq - 0x0101010101010101) & ^eq & 0x8080808080808080) >> 7
		if bitmask != 0 {
			return i + bits.TrailingZeros64(bitmask)
		}
	}

	return -1
}

// Get retrieves a value from the HashMap by its key.
func (h *HashMap[K, V]) Get(key K) (V, bool) {
	hash := h.hashKey(key)
	index := hash & uint64(h.capacity-1)
	hashByte := h.hashToByte(hash)

	for i := uint64(0); i < maxProbeDistance; i++ {
		group := index & ^uint64(groupSize-1)
		match := h.matchGroup(group, hashByte)

		for match != 0 {
			matchIndex := group + uint64(bits.TrailingZeros64(uint64(match)))
			if h.compareKeys(h.entries[matchIndex].key, key) {
				return h.entries[matchIndex].value, true
			}
			match &= match - 1
		}

		if h.ctrl[group] == emptyByte {
			var zero V
			return zero, false
		}

		index = h.nextProbe(index, i)
	}

	var zero V
	return zero, false
}

// nextProbe calculates the next probe index using quadratic probing
func (h *HashMap[K, V]) nextProbe(index, i uint64) uint64 {
	return (index + i*i + i) & uint64(h.capacity-1)
}

// resize increases the capacity of the HashMap and rehashes all elements.
// This is a no-op if the new capacity is less than the current capacity.
// It's not a big deal if we resize a few times, it's still O(1) amortized.
func (h *HashMap[K, V]) resize(newCapacity int) {
	oldCtrl := h.ctrl
	oldEntries := h.entries

	h.capacity = newCapacity
	h.initializeCtrl()
	h.size = 0

	for i, entry := range oldEntries {
		if oldCtrl[i]&0x80 != 0 {
			h.Put(entry.key, entry.value)
		}
	}
}

// hashKey hashes the key using the HashMap's hasher.
func (h *HashMap[K, V]) hashKey(key K) uint64 {
	hash, err := h.hasher.Hash(key)
	if err != nil {
		panic(err)
	}
	return hash
}

// hashToByte converts a hash to a control byte.
// Again, more wicked magic
func (h *HashMap[K, V]) hashToByte(hash uint64) byte {
	return byte((hash >> 57) | 0x80)
}

// compareKeys compares two keys using the HashMap's comparator if available.
func (h *HashMap[K, V]) compareKeys(a, b K) bool {
	if h.comparator != nil {
		return h.comparator(a, b) == 0
	}
	// If no comparator is set, use reflection to compare the keys
	return reflect.DeepEqual(a, b)
}

// Remove removes a key-value pair from the HashMap
// Returns the removed value and a boolean indicating if the key existed.
func (h *HashMap[K, V]) Remove(key K) (V, bool) {
	hash := h.hashKey(key)
	index := hash & uint64(h.capacity-1)
	hashByte := h.hashToByte(hash)

	for i := uint64(0); i < maxProbeDistance; i++ {
		group := index & ^uint64(groupSize-1)
		match := h.matchGroup(group, hashByte)

		for match != 0 {
			matchIndex := group + uint64(bits.TrailingZeros64(uint64(match)))
			if h.compareKeys(h.entries[matchIndex].key, key) {
				return h.removeEntry(matchIndex)
			}
			match &= match - 1
		}

		if h.ctrl[group] == emptyByte {
			var zero V
			return zero, false
		}

		index = h.nextProbe(index, i)
	}

	var zero V
	return zero, false
}

// removeEntry removes an entry at the given index
func (h *HashMap[K, V]) removeEntry(index uint64) (V, bool) {
	removedValue := h.entries[index].value
	h.ctrl[index] = emptyByte
	var zero V
	h.entries[index] = entry[K, V]{key: *new(K), value: zero}
	h.size--
	return removedValue, true
}

// Clear removes all key-value pairs from the HashMap
func (h *HashMap[K, V]) Clear() {
	h.size = 0
	h.capacity = minCapacity
	h.initializeCtrl()
}

// Size returns the number of key-value pairs in the HashMap
func (h *HashMap[K, V]) Size() int {
	return h.size
}

// IsEmpty returns true if the HashMap contains no key-value pairs
func (h *HashMap[K, V]) IsEmpty() bool {
	return h.size == 0
}

// Keys returns a slice containing all the keys in the HashMap
func (h *HashMap[K, V]) Keys() []K {
	keys := make([]K, 0, h.size)
	for i, ctrl := range h.ctrl {
		if ctrl&0x80 != 0 {
			keys = append(keys, h.entries[i].key)
		}
	}
	return keys
}

// Values returns a slice containing all the values in the HashMap
func (h *HashMap[K, V]) Values() []V {
	values := make([]V, 0, h.size)
	for i, ctrl := range h.ctrl {
		if ctrl&0x80 != 0 {
			values = append(values, h.entries[i].value)
		}
	}
	return values
}

// ForEach applies the given function to each key-value pair in the HashMap
func (h *HashMap[K, V]) ForEach(f func(K, V)) {
	for i, ctrl := range h.ctrl {
		if ctrl&0x80 != 0 {
			f(h.entries[i].key, h.entries[i].value)
		}
	}
}

// SetComparator sets a custom comparator for keys
func (h *HashMap[K, V]) SetComparator(comp collections.Comparator[K]) {
	h.comparator = comp
}

// Ensure HashMap implements the Map interface for various types
var (
	_ collections.Map[any, any] = (*HashMap[any, any])(nil)
)

// ContainsKey checks if the given key exists in the HashMap
func (h *HashMap[K, V]) ContainsKey(key K) bool {
	hash := h.hashKey(key)
	index := hash & uint64(h.capacity-1)
	hashByte := h.hashToByte(hash)

	for i := uint64(0); i < maxProbeDistance; i++ {
		group := index & ^uint64(groupSize-1)
		match := h.matchGroup(group, hashByte)

		for match != 0 {
			matchIndex := group + uint64(bits.TrailingZeros64(uint64(match)))
			if h.compareKeys(h.entries[matchIndex].key, key) {
				return true
			}
			match &= match - 1
		}

		if h.ctrl[group] == emptyByte {
			return false
		}

		index = h.nextProbe(index, i)
	}

	return false
}
