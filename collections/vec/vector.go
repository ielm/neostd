package vec

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/errors"
)

// Vec is a contiguous growable array type, similar to Rust's Vec.
type Vec[T any] struct {
	data       []T
	len        int
	cap        int
	comparator comp.Comparator[T]
}

// New creates a new empty Vec without allocating memory.
func New[T any]() *Vec[T] {
	return &Vec[T]{
		data: nil,
		len:  0,
		cap:  0,
	}
}

// VecWithCapacity creates a new Vec with the given capacity and comparator.
func VecWithCapacity[T any](capacity int, comparator comp.Comparator[T]) *Vec[T] {
	return &Vec[T]{
		data:       make([]T, 0, capacity),
		len:        0,
		cap:        capacity,
		comparator: comparator,
	}
}

// Push appends an element to the back of the Vec.
func (v *Vec[T]) Push(item T) {
	if v.len == v.cap {
		v.grow()
	}
	if v.cap == 0 {
		v.data = make([]T, 1)
	} else {
		v.data = v.data[:v.len+1]
	}
	v.data[v.len] = item
	v.len++
}

// Pop removes and returns the last element from the Vec.
// If the Vec is empty, it returns the zero value of T and false.
func (v *Vec[T]) Pop() (T, bool) {
	if v.len == 0 {
		var zero T
		return zero, false
	}
	item := v.data[v.len-1]
	v.data = v.data[:v.len-1]
	v.len--
	return item, true
}

// Get returns the element at the given index.
// If the index is out of bounds, it returns the zero value of T and an error.
func (v *Vec[T]) Get(index int) (T, error) {
	if index < 0 || index >= v.len {
		var zero T
		return zero, errors.New(errors.ErrOutOfBounds, "index out of bounds")
	}
	return v.data[index], nil
}

// Set sets the element at the given index.
// If the index is out of bounds, it returns an error.
func (v *Vec[T]) Set(index int, item T) error {
	if index < 0 || index >= v.len {
		return errors.New(errors.ErrOutOfBounds, "index out of bounds")
	}
	v.data[index] = item
	return nil
}

// Len returns the number of elements in the Vec.
func (v *Vec[T]) Len() int {
	return v.len
}

// Cap returns the capacity of the Vec.
func (v *Vec[T]) Cap() int {
	return v.cap
}

// Clear removes all elements from the Vec.
func (v *Vec[T]) Clear() {
	v.data = v.data[:0]
	v.len = 0
}

// IsEmpty returns true if the Vec contains no elements.
func (v *Vec[T]) IsEmpty() bool {
	return v.len == 0
}

// grow increases the capacity of the Vec.
func (v *Vec[T]) grow() {
	newCap := v.cap * 2
	if newCap == 0 {
		newCap = 1
	}
	v.Grow(newCap)
}

// Grow increases the capacity of the Vec to the specified size.
func (v *Vec[T]) Grow(newCap int) {
	if newCap > v.cap {
		newData := make([]T, v.len, newCap)
		if v.data != nil {
			copy(newData, v.data)
		}
		v.data = newData
		v.cap = newCap
	}
}

// SetComparator sets the comparator for the Vec.
func (v *Vec[T]) SetComparator(comparator comp.Comparator[T]) {
	v.comparator = comparator
}

// Contains checks if the Vec contains the given item.
func (v *Vec[T]) Contains(item T) bool {
	if v.comparator == nil {
		panic("comparator not set for non-comparable type")
	}
	for _, elem := range v.data[:v.len] {
		if v.comparator(elem, item) == 0 {
			return true
		}
	}
	return false
}

// IndexOf returns the index of the first occurrence of the given item.
// If the item is not found, it returns -1.
func (v *Vec[T]) IndexOf(item T) int {
	if v.comparator == nil {
		panic("comparator not set for non-comparable type")
	}
	for i, elem := range v.data[:v.len] {
		if v.comparator(elem, item) == 0 {
			return i
		}
	}
	return -1
}

// Remove removes the first occurrence of the given item from the Vec.
// It returns true if the item was found and removed, false otherwise.
func (v *Vec[T]) Remove(item T) bool {
	index := v.IndexOf(item)
	if index == -1 {
		return false
	}
	v.RemoveAt(index)
	return true
}

// RemoveAt removes the element at the given index.
// If the index is out of bounds, it returns an error.
func (v *Vec[T]) RemoveAt(index int) error {
	if index < 0 || index >= v.len {
		return errors.New(errors.ErrOutOfBounds, "index out of bounds")
	}
	copy(v.data[index:], v.data[index+1:])
	v.len--
	v.data[v.len] = *new(T) // zero the last element
	return nil
}

// Iterator returns an iterator for the Vec.
func (v *Vec[T]) Iterator() collections.Iterator[T] {
	return &vecIterator[T]{vec: v, index: 0}
}

// ReverseIterator returns a reverse iterator for the Vec.
func (v *Vec[T]) ReverseIterator() collections.Iterator[T] {
	return &vecReverseIterator[T]{vec: v, index: v.len - 1}
}

type vecIterator[T any] struct {
	vec   *Vec[T]
	index int
}

func (it *vecIterator[T]) HasNext() bool {
	return it.index < it.vec.len
}

func (it *vecIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	item := it.vec.data[it.index]
	it.index++
	return item
}

type vecReverseIterator[T any] struct {
	vec   *Vec[T]
	index int
}

func (it *vecReverseIterator[T]) HasNext() bool {
	return it.index >= 0
}

func (it *vecReverseIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	item := it.vec.data[it.index]
	it.index--
	return item
}

// Add implements the Collection interface.
func (v *Vec[T]) Add(item T) bool {
	v.Push(item)
	return true
}

// Size implements the Collection interface.
func (v *Vec[T]) Size() int {
	return v.len
}

// Ensure Vec implements the Vector interface
var _ collections.Vector[any] = (*Vec[any])(nil)
