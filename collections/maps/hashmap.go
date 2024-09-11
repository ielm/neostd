package maps

import (
	"math/bits"
	"sync"
	"unsafe"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/hash"
	"github.com/ielm/neostd/res"
)

// Constants
const (
	defaultLoadFactor = 0.875
	minCapacity       = 8
	groupSize         = 16
	maxProbeDistance  = 128
	emptyByte         = 0b11111111
)

// HashMap struct definition
type HashMap[K any, V any] struct {
	mu         sync.RWMutex // Add a read-write mutex for thread safety
	ctrl       []byte
	entries    []entry[K, V]
	size       int
	capacity   int
	loadFactor float64
	hasher     hash.Hasher
	comparator comp.Comparator[K]
}

// entry struct definition
type entry[K any, V any] struct {
	key   K
	value V
}

// NewHashMap creates a new HashMap with default settings.
// It initializes the map with a minimum capacity and default load factor.
//
// The comparator parameter is used for key comparison. For built-in types,
// you can use collections.GenericComparator[K]().
//
// Example:
//
//	hm := maps.NewHashMap[string, int](collections.GenericComparator[string]())
func NewHashMap[K any, V any](comparator comp.Comparator[K]) res.Result[*HashMap[K, V]] {
	hasher, err := hash.NewSipHasher()
	if err != nil {
		return res.Err[*HashMap[K, V]](err)
	}
	return NewHashMapWithHasher[K, V](comparator, hasher)
}

// NewHashMapWithHasher creates a new HashMap with a custom hasher.
// This allows for more flexibility in how keys are hashed.
//
// Example:
//
//	customHasher := &MyCustomHasher{}
//	hm := maps.NewHashMapWithHasher[string, int](collections.GenericComparator[string](), customHasher)
func NewHashMapWithHasher[K any, V any](comparator comp.Comparator[K], hasher hash.Hasher) res.Result[*HashMap[K, V]] {
	h := &HashMap[K, V]{
		capacity:   minCapacity,
		loadFactor: defaultLoadFactor,
		comparator: comparator,
		hasher:     hasher,
	}
	h.initializeCtrl()
	return res.Ok(h)
}

// Core methods

// Put inserts a key-value pair into the HashMap.
// If the key already exists, the old value is replaced and returned.
// The boolean return value indicates whether an existing entry was updated.
//
// Example:
//
//	oldValue, existed := hm.Put("key", 42)
func (h *HashMap[K, V]) Put(key K, value V) (V, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

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

// Get retrieves a value from the HashMap by its key.
// It returns the value and a boolean indicating whether the key was found.
// TODO: Use Option[V]
//
// Example:
//
//	value, found := hm.Get("key")
func (h *HashMap[K, V]) Get(key K) (V, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

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

// Remove removes a key-value pair from the HashMap
// Returns the removed value and a boolean indicating if the key existed.
//
// Example:
//
//	removedValue, existed := hm.Remove("key")
func (h *HashMap[K, V]) Remove(key K) (V, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

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

// Helper methods

// initializeCtrl initializes the control bytes and entries.
func (h *HashMap[K, V]) initializeCtrl() {
	h.ctrl = make([]byte, h.capacity+groupSize)
	for i := range h.ctrl {
		h.ctrl[i] = emptyByte
	}
	h.entries = make([]entry[K, V], h.capacity)
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
	for i := 0; i < 16; i += 8 {
		// Load 8 bytes from the vector
		chunk := *(*uint64)(unsafe.Pointer(&vec[i]))
		// XOR the chunk with the hashByte
		eq := chunk ^ (uint64(hashByte) * 0x0101010101010101)
		// This is a bitmask that will be non-zero if the hashByte is found in the chunk
		bitmask := ((eq - 0x0101010101010101) & ^eq & 0x8080808080808080) >> 7
		// OR the bitmask with the mask
		mask |= uint16(bitmask) << i
	}

	return mask
}

// findEmptySlot finds an empty slot in a group.
func (h *HashMap[K, V]) findEmptySlot(group uint64) int {
	vec := (*[16]uint8)(unsafe.Pointer(&h.ctrl[group]))

	// Check 16 slots in parallel
	for i := 0; i < 16; i += 8 {
		// Load 8 bytes from the vector
		chunk := *(*uint64)(unsafe.Pointer(&vec[i]))
		// XOR the chunk with the hashByte
		eq := chunk ^ (uint64(emptyByte) * 0x0101010101010101)
		// Bitmask to check if any of the 8 bytes are empty
		bitmask := ((eq - 0x0101010101010101) & ^eq & 0x8080808080808080) >> 7
		// If the bitmask is not zero, we found an empty slot
		if bitmask != 0 {
			return i + bits.TrailingZeros64(bitmask)
		}
	}

	return -1
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
	keyBytes, err := keyToBytes(key)
	if err != nil {
		panic(err) // In production, consider handling this error more gracefully
	}
	h.hasher.Reset()
	h.hasher.Write(keyBytes)
	hashBytes := h.hasher.Sum(nil)
	return hash.HashBytesToUint64(hashBytes)
}

// hashToByte converts a hash to a control byte.
// Again, more wicked magic
func (h *HashMap[K, V]) hashToByte(hash uint64) byte {
	return byte((hash >> 57) | 0x80)
}

// compareKeys compares two keys using the HashMap's comparator.
func (h *HashMap[K, V]) compareKeys(a, b K) bool {
	return h.comparator(a, b) == 0
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

// Interface Compliance methods

// Clear removes all key-value pairs from the HashMap
//
// Example:
//
//	hm.Clear()
func (h *HashMap[K, V]) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.size = 0
	h.capacity = minCapacity
	h.initializeCtrl()
}

// Size returns the number of key-value pairs in the HashMap
//
// Example:
//
//	count := hm.Size()
func (h *HashMap[K, V]) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.size
}

// IsEmpty returns true if the HashMap contains no key-value pairs
//
// Example:
//
//	if hm.IsEmpty() {
//		fmt.Println("HashMap is empty")
//	}
func (h *HashMap[K, V]) IsEmpty() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.size == 0
}

// Keys returns a slice containing all the keys in the HashMap
//
// Example:
//
//	keys := hm.Keys()
func (h *HashMap[K, V]) Keys() []K {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keys := make([]K, 0, h.size)
	for i, ctrl := range h.ctrl {
		if ctrl&0x80 != 0 {
			keys = append(keys, h.entries[i].key)
		}
	}
	return keys
}

// Values returns a slice containing all the values in the HashMap
//
// Example:
//
//	values := hm.Values()
func (h *HashMap[K, V]) Values() []V {
	h.mu.RLock()
	defer h.mu.RUnlock()

	values := make([]V, 0, h.size)
	for i, ctrl := range h.ctrl {
		if ctrl&0x80 != 0 {
			values = append(values, h.entries[i].value)
		}
	}
	return values
}

// ForEach applies the given function to each key-value pair in the HashMap
//
// Example:
//
//	hm.ForEach(func(key string, value int) {
//		fmt.Printf("Key: %s, Value: %d\n", key, value)
//	})
func (h *HashMap[K, V]) ForEach(f func(K, V)) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i, ctrl := range h.ctrl {
		if ctrl&0x80 != 0 {
			f(h.entries[i].key, h.entries[i].value)
		}
	}
}

// ContainsKey checks if the given key exists in the HashMap
func (h *HashMap[K, V]) ContainsKey(key K) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, found := h.Get(key)
	return found
}

// SetComparator sets a custom comparator for keys
// This is particularly useful for complex key types or when you need specific comparison logic.
//
// Example:
//
//	hm.SetComparator(func(a, b MyKeyType) int {
//		// Custom comparison logic
//		return a.CompareTo(b)
//	})
func (h *HashMap[K, V]) SetComparator(comp comp.Comparator[K]) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.comparator = comp
}

// Comparator returns the comparator for the HashMap.
func (h *HashMap[K, V]) Comparator() comp.Comparator[K] {
	return h.comparator
}

// Type assertions
var (
	_ collections.Map[string, any]      = (*HashMap[string, any])(nil)
	_ collections.Map[int, any]         = (*HashMap[int, any])(nil)
	_ collections.Map[bool, any]        = (*HashMap[bool, any])(nil)
	_ collections.Map[interface{}, any] = (*HashMap[interface{}, any])(nil)
)

// T is an example of a type that's not inherently comparable
type T interface{}

// Ensure HashMap implements the Map interface for T
var _ collections.Map[T, any] = (*HashMap[T, any])(nil)

// keyToBytes converts a key of any type to a byte slice
func keyToBytes(key any) ([]byte, error) {
	switch k := key.(type) {
	case string:
		return []byte(k), nil
	case []byte:
		return k, nil
	default:
		return hash.ToBinary(k)
	}
}
