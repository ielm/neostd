package list

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/hash"
)

// Constants for SkipList configuration
const (
	maxLevel    = 32   // Maximum number of levels in the SkipList
	probability = 0.25 // Probability of promoting a node to the next level
)

// SkipList is a highly optimized, concurrent-safe implementation of a skip list.
//
// A SkipList is a probabilistic data structure that allows for fast search, insertion, and deletion
// operations. It maintains multiple layers of linked lists, with each higher layer acting as an
// "express lane" for the layers below, allowing for O(log n) average time complexity for these operations.
//
// The structure consists of nodes, where each node contains a value and multiple forward pointers.
// The number of forward pointers (the node's level) is randomly determined during insertion,
// following a probability distribution that ensures a balanced structure.
//
// This implementation is thread-safe and uses a comparator for ordering elements and a hasher for
// generating hash values when needed.
type SkipList[T any] struct {
	head   *node[T]           // Pointer to the head (sentinel) node
	tail   *node[T]           // Pointer to the tail (sentinel) node
	length int                // Number of elements in the SkipList
	level  int                // Current maximum level of the SkipList
	comp   comp.Comparator[T] // Comparator function for ordering elements
	hasher hash.Hasher        // Hasher for generating hash values
	mu     sync.RWMutex       // Read-write mutex for concurrent access
}

// node represents a single element in the SkipList
type node[T any] struct {
	value    T          // The value stored in the node
	forward  []*node[T] // Array of forward pointers to next nodes at each level
	backward *node[T]   // Pointer to the previous node (for reverse iteration)
}

// NewWithHasher creates a new SkipList with the given comparator and hasher.
//
// The comparator is used for ordering elements, while the hasher is used for
// generating hash values when needed (e.g., for certain operations or extensions).
//
// Example:
//
//	comp := collections.GenericComparator[int]()
//	hasher, _ := hash.NewSipHasher()
//	sl, err := NewWithHasher(comp, hasher)
func NewWithHasher[T any](comp comp.Comparator[T], hasher hash.Hasher) (*SkipList[T], error) {
	sl := &SkipList[T]{
		level:  1,
		comp:   comp,
		hasher: hasher,
	}
	sl.head = sl.newNode(maxLevel, *new(T))
	sl.tail = sl.newNode(maxLevel, *new(T))
	for i := 0; i < maxLevel; i++ {
		sl.head.forward[i] = sl.tail
	}
	sl.tail.backward = sl.head
	return sl, nil
}

// NewSkipList creates a new SkipList with the given comparator and a default hasher.
//
// This is a convenience function that uses a default SipHasher. For more control
// over the hasher, use NewWithHasher instead.
//
// Example:
//
//	comp := collections.GenericComparator[string]()
//	sl, err := NewSkipList(comp)
func NewSkipList[T any](comp comp.Comparator[T]) (*SkipList[T], error) {
	hasher, err := hash.NewSipHasher()
	if err != nil {
		return nil, fmt.Errorf("failed to create hasher: %w", err)
	}
	return NewWithHasher(comp, hasher)
}

// newNode creates a new node with the given level and value
func (sl *SkipList[T]) newNode(level int, value T) *node[T] {
	return &node[T]{
		value:   value,
		forward: make([]*node[T], level),
	}
}

// Insert adds an element to the SkipList.
//
// This method maintains the order of elements based on the comparator.
// If the element already exists, it will be inserted after the existing elements
// with the same value.
//
// Example:
//
//	sl.Insert(42)
func (sl *SkipList[T]) Insert(value T) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*node[T], maxLevel)
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.forward[i] != sl.tail && sl.comp(x.forward[i].value, value) < 0 {
			x = x.forward[i]
		}
		update[i] = x
	}

	level := sl.randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	newNode := sl.newNode(level, value)
	for i := 0; i < level; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	newNode.backward = update[0]
	if newNode.forward[0] != sl.tail {
		newNode.forward[0].backward = newNode
	} else {
		sl.tail.backward = newNode
	}

	sl.length++
}

// Remove deletes an element from the SkipList.
//
// This method removes the first occurrence of the specified value.
// It returns true if the element was found and removed, false otherwise.
//
// Example:
//
//	removed := sl.Remove(42)
func (sl *SkipList[T]) Remove(value T) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*node[T], maxLevel)
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.forward[i] != sl.tail && sl.comp(x.forward[i].value, value) < 0 {
			x = x.forward[i]
		}
		update[i] = x
	}

	x = x.forward[0]
	if x != sl.tail && sl.comp(x.value, value) == 0 {
		for i := 0; i < sl.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}

		if x.forward[0] != sl.tail {
			x.forward[0].backward = update[0]
		} else {
			sl.tail.backward = update[0]
		}

		for sl.level > 1 && sl.head.forward[sl.level-1] == sl.tail {
			sl.level--
		}

		sl.length--
		return true
	}

	return false
}

// Contains checks if an element exists in the SkipList.
//
// This method uses the comparator to check for equality.
// It returns true if the element is found, false otherwise.
//
// Example:
//
//	exists := sl.Contains(42)
func (sl *SkipList[T]) Contains(value T) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for x.forward[i] != sl.tail && sl.comp(x.forward[i].value, value) < 0 {
			x = x.forward[i]
		}
	}

	x = x.forward[0]
	return x != sl.tail && sl.comp(x.value, value) == 0
}

// Size returns the number of elements in the SkipList.
//
// Example:
//
//	count := sl.Size()
func (sl *SkipList[T]) Size() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}

// IsEmpty returns true if the SkipList is empty.
//
// Example:
//
//	if sl.IsEmpty() {
//		fmt.Println("SkipList is empty")
//	}
func (sl *SkipList[T]) IsEmpty() bool {
	return sl.Size() == 0
}

// Clear removes all elements from the SkipList.
//
// This method resets the SkipList to its initial state.
//
// Example:
//
//	sl.Clear()
func (sl *SkipList[T]) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.head = sl.newNode(maxLevel, *new(T))
	sl.tail = sl.newNode(maxLevel, *new(T))
	for i := 0; i < maxLevel; i++ {
		sl.head.forward[i] = sl.tail
	}
	sl.tail.backward = sl.head
	sl.length = 0
	sl.level = 1
}

// Iterator returns an iterator for the SkipList.
//
// The iterator allows forward traversal of the SkipList elements.
//
// Example:
//
//	it := sl.Iterator()
//	for it.HasNext() {
//		fmt.Println(it.Next())
//	}
func (sl *SkipList[T]) Iterator() collections.Iterator[T] {
	return &skipListIterator[T]{
		current: sl.head.forward[0],
		tail:    sl.tail,
	}
}

// ReverseIterator returns a reverse iterator for the SkipList.
//
// The reverse iterator allows backward traversal of the SkipList elements.
//
// Example:
//
//	it := sl.ReverseIterator()
//	for it.HasNext() {
//		fmt.Println(it.Next())
//	}
func (sl *SkipList[T]) ReverseIterator() collections.Iterator[T] {
	return &skipListReverseIterator[T]{
		current: sl.tail.backward,
		head:    sl.head,
	}
}

type skipListIterator[T any] struct {
	current *node[T]
	tail    *node[T]
}

func (it *skipListIterator[T]) HasNext() bool {
	return it.current != it.tail
}

func (it *skipListIterator[T]) Next() T {
	if !it.HasNext() {
		panic("SkipListIterator: No more elements")
	}
	value := it.current.value
	it.current = it.current.forward[0]
	return value
}

type skipListReverseIterator[T any] struct {
	current *node[T]
	head    *node[T]
}

func (it *skipListReverseIterator[T]) HasNext() bool {
	return it.current != it.head
}

func (it *skipListReverseIterator[T]) Next() T {
	if !it.HasNext() {
		panic("SkipListReverseIterator: No more elements")
	}
	value := it.current.value
	it.current = it.current.backward
	return value
}

// randomLevel generates a random level for a new node.
//
// This method uses a probabilistic approach to determine the level of a new node,
// which is crucial for maintaining the SkipList's balance and performance characteristics.
func (sl *SkipList[T]) randomLevel() int {
	level := 1
	for fastrand() < uint32(float32(probability)*math.MaxUint32) && level < maxLevel {
		level++
	}
	return level
}

// fastrand is a fast, thread-safe random number generator.
func fastrand() uint32 {
	return uint32(uintptr(unsafe.Pointer(new(byte))))*1664525 + 1013904223
}

// Get retrieves an element from the SkipList by its value.
//
// This method returns the value and a boolean indicating whether the value was found.
//
// Example:
//
//	value, found := sl.Get(42)
func (sl *SkipList[T]) Get(value T) (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for x.forward[i] != sl.tail && sl.comp(x.forward[i].value, value) < 0 {
			x = x.forward[i]
		}
	}

	x = x.forward[0]
	if x != sl.tail && sl.comp(x.value, value) == 0 {
		return x.value, true
	}

	return *new(T), false
}

// Add an element to the SkipList (to satisfy the Set interface)
//
// This method is an alias for Insert to conform to the Set interface.
//
// Example:
//
//	added := sl.Add(42)
func (sl *SkipList[T]) Add(value T) bool {
	sl.Insert(value)
	return true
}

// SetComparator sets the comparator for the SkipList
//
// This method allows changing the comparison logic for the SkipList elements.
// Note that changing the comparator on a non-empty SkipList may lead to inconsistencies.
//
// Example:
//
//	sl.SetComparator(func(a, b int) int {
//		return b - a // Reverse order
//	})
func (sl *SkipList[T]) SetComparator(comp comp.Comparator[T]) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.comp = comp
}

// Ensure SkipList implements the SortedSet interface
var _ collections.SortedSet[any] = (*SkipList[any])(nil)
