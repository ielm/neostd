package set

import (
	"fmt"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
)

// DisjointSet represents a disjoint-set data structure.
//
// A DisjointSet, also known as a union-find data structure, maintains a collection
// of disjoint (non-overlapping) sets. It provides near-constant-time operations
// to add new sets, merge existing sets, and determine whether elements are in the same set.
//
// This implementation uses both union by rank and path compression optimizations
// to achieve near-constant time complexity for most operations.
type DisjointSet[T comparable] struct {
	parent   map[T]T   // Maps each element to its parent
	rank     map[T]int // Stores the rank (approximate depth) of each tree
	size     map[T]int // Stores the size of each set
	setCount int       // The number of disjoint sets
}

// NewDisjointSet creates a new DisjointSet.
//
// Example:
//
//	ds := NewDisjointSet[int]()
func NewDisjointSet[T comparable]() *DisjointSet[T] {
	return &DisjointSet[T]{
		parent:   make(map[T]T),
		rank:     make(map[T]int),
		size:     make(map[T]int),
		setCount: 0,
	}
}

// MakeSet creates a new set containing only the given item.
//
// Returns true if a new set was created, false if the item already existed.
//
// Example:
//
//	ds.MakeSet(1)
//	ds.MakeSet(2)
func (ds *DisjointSet[T]) MakeSet(item T) bool {
	if _, exists := ds.parent[item]; exists {
		return false
	}
	ds.parent[item] = item
	ds.rank[item] = 0
	ds.size[item] = 1
	ds.setCount++
	return true
}

// Find returns the representative (root) of the set containing the given item.
// It uses path compression for optimization.
//
// Example:
//
//	root, err := ds.Find(1)
func (ds *DisjointSet[T]) Find(item T) (T, error) {
	if _, exists := ds.parent[item]; !exists {
		var zero T
		return zero, fmt.Errorf("item %v not found in any set", item)
	}

	if ds.parent[item] != item {
		root, _ := ds.Find(ds.parent[item])
		ds.parent[item] = root // Path compression
	}
	return ds.parent[item], nil
}

// Union merges the sets containing items x and y.
// It uses union by rank for optimization.
//
// Example:
//
//	err := ds.Union(1, 2)
func (ds *DisjointSet[T]) Union(x, y T) error {
	rootX, errX := ds.Find(x)
	rootY, errY := ds.Find(y)

	if errX != nil {
		return errX
	}
	if errY != nil {
		return errY
	}

	if rootX == rootY {
		return nil // Already in the same set
	}

	// Union by rank
	if ds.rank[rootX] < ds.rank[rootY] {
		ds.parent[rootX] = rootY
		ds.size[rootY] += ds.size[rootX]
	} else if ds.rank[rootX] > ds.rank[rootY] {
		ds.parent[rootY] = rootX
		ds.size[rootX] += ds.size[rootY]
	} else {
		ds.parent[rootY] = rootX
		ds.rank[rootX]++
		ds.size[rootX] += ds.size[rootY]
	}

	ds.setCount--
	return nil
}

// Connected checks if two items are in the same set.
//
// Example:
//
//	connected, err := ds.Connected(1, 2)
func (ds *DisjointSet[T]) Connected(x, y T) (bool, error) {
	rootX, errX := ds.Find(x)
	rootY, errY := ds.Find(y)

	if errX != nil {
		return false, errX
	}
	if errY != nil {
		return false, errY
	}

	return rootX == rootY, nil
}

// SetSize returns the size of the set containing the given item.
//
// Example:
//
//	size, err := ds.SetSize(1)
func (ds *DisjointSet[T]) SetSize(item T) (int, error) {
	root, err := ds.Find(item)
	if err != nil {
		return 0, err
	}
	return ds.size[root], nil
}

// SetCount returns the number of disjoint sets.
//
// Example:
//
//	count := ds.SetCount()
func (ds *DisjointSet[T]) SetCount() int {
	return ds.setCount
}

// Clear removes all elements from the DisjointSet.
//
// Example:
//
//	ds.Clear()
func (ds *DisjointSet[T]) Clear() {
	ds.parent = make(map[T]T)
	ds.rank = make(map[T]int)
	ds.size = make(map[T]int)
	ds.setCount = 0
}

// IsEmpty returns true if the DisjointSet contains no elements.
//
// Example:
//
//	if ds.IsEmpty() {
//		fmt.Println("DisjointSet is empty")
//	}
func (ds *DisjointSet[T]) IsEmpty() bool {
	return len(ds.parent) == 0
}

// Size returns the total number of elements in the DisjointSet.
//
// Example:
//
//	totalElements := ds.Size()
func (ds *DisjointSet[T]) Size() int {
	return len(ds.parent)
}

// Contains checks if the given item exists in any set.
//
// Example:
//
//	if ds.Contains(1) {
//		fmt.Println("Item 1 exists in the DisjointSet")
//	}
func (ds *DisjointSet[T]) Contains(item T) bool {
	_, exists := ds.parent[item]
	return exists
}

// Add adds a new item to the DisjointSet in its own set.
// This is an alias for MakeSet to satisfy the Set interface.
//
// Example:
//
//	added := ds.Add(3)
func (ds *DisjointSet[T]) Add(item T) bool {
	return ds.MakeSet(item)
}

// Remove removes an item from the DisjointSet.
//
// Example:
//
//	removed := ds.Remove(1)
func (ds *DisjointSet[T]) Remove(item T) bool {
	if !ds.Contains(item) {
		return false
	}

	root, _ := ds.Find(item)
	if root == item {
		// Item is a root, need to update all its children
		for child, parent := range ds.parent {
			if parent == item && child != item {
				ds.parent[child] = child
				ds.rank[child] = 0
				ds.size[child] = 1
				ds.setCount++
			}
		}
	}

	delete(ds.parent, item)
	delete(ds.rank, item)
	delete(ds.size, item)

	if root == item {
		ds.setCount--
	} else {
		ds.size[root]--
	}

	return true
}

// SetComparator is a no-op for DisjointSet as it doesn't use comparators.
func (ds *DisjointSet[T]) SetComparator(comp.Comparator[T]) {
	// No-op
}

// Iterator returns an iterator over all elements in the DisjointSet.
//
// Example:
//
//	it := ds.Iterator()
//	for it.HasNext() {
//		fmt.Println(it.Next())
//	}
func (ds *DisjointSet[T]) Iterator() collections.Iterator[T] {
	items := make([]T, 0, len(ds.parent))
	for item := range ds.parent {
		items = append(items, item)
	}
	return &disjointSetIterator[T]{items: items, index: 0}
}

// ReverseIterator returns a reverse iterator over all elements in the DisjointSet.
//
// Example:
//
//	it := ds.ReverseIterator()
//	for it.HasNext() {
//		fmt.Println(it.Next())
//	}
func (ds *DisjointSet[T]) ReverseIterator() collections.Iterator[T] {
	items := make([]T, 0, len(ds.parent))
	for item := range ds.parent {
		items = append(items, item)
	}
	return &disjointSetIterator[T]{items: items, index: len(items) - 1, reverse: true}
}

type disjointSetIterator[T comparable] struct {
	items   []T
	index   int
	reverse bool
}

func (it *disjointSetIterator[T]) HasNext() bool {
	if it.reverse {
		return it.index >= 0
	}
	return it.index < len(it.items)
}

func (it *disjointSetIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	item := it.items[it.index]
	if it.reverse {
		it.index--
	} else {
		it.index++
	}
	return item
}

// Ensure DisjointSet implements the Set interface
var _ collections.Set[int] = (*DisjointSet[int])(nil)
