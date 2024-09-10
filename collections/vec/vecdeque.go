package vec

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/errors"
)

// VecDeque is a double-ended queue implemented with a growable ring buffer.
type VecDeque[T any] struct {
	buf        []T
	head       int
	tail       int
	len        int
	cap        int
	comparator comp.Comparator[T]
}

// NewVecDeque creates a new VecDeque with the given capacity.
func NewVecDeque[T any](capacity int) *VecDeque[T] {
	return &VecDeque[T]{
		buf:  make([]T, capacity),
		cap:  capacity,
		head: 0,
		tail: 0,
		len:  0,
	}
}

// VecDequeWithCapacity creates a new VecDeque with the given capacity and comparator.
func VecDequeWithCapacity[T any](capacity int, comparator comp.Comparator[T]) *VecDeque[T] {
	return &VecDeque[T]{
		buf:        make([]T, capacity),
		cap:        capacity,
		head:       0,
		tail:       0,
		len:        0,
		comparator: comparator,
	}
}

// PushBack appends an element to the back of the VecDeque.
func (vd *VecDeque[T]) PushBack(item T) {
	if vd.len == vd.cap {
		vd.grow()
	}
	vd.buf[vd.tail] = item
	vd.tail = (vd.tail + 1) % vd.cap
	vd.len++
}

// PushFront prepends an element to the front of the VecDeque.
func (vd *VecDeque[T]) PushFront(item T) {
	if vd.len == vd.cap {
		vd.grow()
	}
	vd.head = (vd.head - 1 + vd.cap) % vd.cap
	vd.buf[vd.head] = item
	vd.len++
}

// PopBack removes and returns the last element from the VecDeque.
// If the VecDeque is empty, it returns the zero value of T and false.
func (vd *VecDeque[T]) PopBack() (T, bool) {
	if vd.IsEmpty() {
		var zero T
		return zero, false
	}
	vd.tail = (vd.tail - 1 + vd.cap) % vd.cap
	item := vd.buf[vd.tail]
	vd.len--
	return item, true
}

// PopFront removes and returns the first element from the VecDeque.
// If the VecDeque is empty, it returns the zero value of T and false.
func (vd *VecDeque[T]) PopFront() (T, bool) {
	if vd.IsEmpty() {
		var zero T
		return zero, false
	}
	item := vd.buf[vd.head]
	vd.head = (vd.head + 1) % vd.cap
	vd.len--
	return item, true
}

// Front returns the first element of the VecDeque without removing it.
// If the VecDeque is empty, it returns the zero value of T and false.
func (vd *VecDeque[T]) Front() (T, bool) {
	if vd.IsEmpty() {
		var zero T
		return zero, false
	}
	return vd.buf[vd.head], true
}

// Back returns the last element of the VecDeque without removing it.
// If the VecDeque is empty, it returns the zero value of T and false.
func (vd *VecDeque[T]) Back() (T, bool) {
	if vd.IsEmpty() {
		var zero T
		return zero, false
	}
	return vd.buf[(vd.tail-1+vd.cap)%vd.cap], true
}

// Get returns the element at the given index.
// If the index is out of bounds, it returns the zero value of T and an error.
func (vd *VecDeque[T]) Get(index int) (T, error) {
	if index < 0 || index >= vd.len {
		var zero T
		return zero, errors.New(errors.ErrOutOfBounds, "index out of bounds")
	}
	return vd.buf[(vd.head+index)%vd.cap], nil
}

// Set sets the element at the given index.
// If the index is out of bounds, it returns an error.
func (vd *VecDeque[T]) Set(index int, item T) error {
	if index < 0 || index >= vd.len {
		return errors.New(errors.ErrOutOfBounds, "index out of bounds")
	}
	vd.buf[(vd.head+index)%vd.cap] = item
	return nil
}

// Len returns the number of elements in the VecDeque.
func (vd *VecDeque[T]) Len() int {
	return vd.len
}

// Cap returns the capacity of the VecDeque.
func (vd *VecDeque[T]) Cap() int {
	return vd.cap
}

// Clear removes all elements from the VecDeque.
func (vd *VecDeque[T]) Clear() {
	vd.head = 0
	vd.tail = 0
	vd.len = 0
}

// IsEmpty returns true if the VecDeque contains no elements.
func (vd *VecDeque[T]) IsEmpty() bool {
	return vd.len == 0
}

// grow increases the capacity of the VecDeque.
func (vd *VecDeque[T]) grow() {
	newCap := vd.cap * 2
	if newCap == 0 {
		newCap = 1
	}
	vd.Grow(newCap)
}

// Grow increases the capacity of the VecDeque to the specified size.
func (vd *VecDeque[T]) Grow(newCap int) {
	if newCap > vd.cap {
		newBuf := make([]T, newCap)
		if vd.tail > vd.head {
			copy(newBuf, vd.buf[vd.head:vd.tail])
		} else {
			n := copy(newBuf, vd.buf[vd.head:])
			copy(newBuf[n:], vd.buf[:vd.tail])
		}
		vd.buf = newBuf
		vd.head = 0
		vd.tail = vd.len
		vd.cap = newCap
	}
}

// SetComparator sets the comparator for the VecDeque.
func (vd *VecDeque[T]) SetComparator(comparator comp.Comparator[T]) {
	vd.comparator = comparator
}

// Contains checks if the VecDeque contains the given item.
func (vd *VecDeque[T]) Contains(item T) bool {
	if vd.comparator == nil {
		panic("comparator not set for non-comparable type")
	}
	for i := 0; i < vd.len; i++ {
		if vd.comparator(vd.buf[(vd.head+i)%vd.cap], item) == 0 {
			return true
		}
	}
	return false
}

// IndexOf returns the index of the first occurrence of the given item.
// If the item is not found, it returns -1.
func (vd *VecDeque[T]) IndexOf(item T) int {
	if vd.comparator == nil {
		panic("comparator not set for non-comparable type")
	}
	for i := 0; i < vd.len; i++ {
		if vd.comparator(vd.buf[(vd.head+i)%vd.cap], item) == 0 {
			return i
		}
	}
	return -1
}

// Remove removes the first occurrence of the given item from the VecDeque.
// It returns true if the item was found and removed, false otherwise.
func (vd *VecDeque[T]) Remove(item T) bool {
	index := vd.IndexOf(item)
	if index == -1 {
		return false
	}
	vd.RemoveAt(index)
	return true
}

// RemoveAt removes the element at the given index.
// If the index is out of bounds, it returns an error.
func (vd *VecDeque[T]) RemoveAt(index int) error {
	if index < 0 || index >= vd.len {
		return errors.New(errors.ErrOutOfBounds, "index out of bounds")
	}
	if index == 0 {
		vd.PopFront()
	} else if index == vd.len-1 {
		vd.PopBack()
	} else {
		for i := index; i < vd.len-1; i++ {
			vd.Set(i, vd.buf[(vd.head+i+1)%vd.cap])
		}
		vd.tail = (vd.tail - 1 + vd.cap) % vd.cap
		vd.len--
	}
	return nil
}

// MakeContiguous rotates the VecDeque so that its elements do not wrap,
// and returns a mutable slice to the now-contiguous element sequence.
func (vd *VecDeque[T]) MakeContiguous() []T {
	if vd.head <= vd.tail {
		return vd.buf[vd.head:vd.tail]
	}
	vd.buf = append(vd.buf[vd.head:], vd.buf[:vd.tail]...)
	vd.head = 0
	vd.tail = vd.len
	return vd.buf[:vd.len]
}

// Iterator returns an iterator for the VecDeque.
func (vd *VecDeque[T]) Iterator() collections.Iterator[T] {
	return &vecDequeIterator[T]{vd: vd, index: 0}
}

// ReverseIterator returns a reverse iterator for the VecDeque.
func (vd *VecDeque[T]) ReverseIterator() collections.Iterator[T] {
	return &vecDequeReverseIterator[T]{vd: vd, index: vd.len - 1}
}

type vecDequeIterator[T any] struct {
	vd    *VecDeque[T]
	index int
}

func (it *vecDequeIterator[T]) HasNext() bool {
	return it.index < it.vd.len
}

func (it *vecDequeIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	item, _ := it.vd.Get(it.index)
	it.index++
	return item
}

type vecDequeReverseIterator[T any] struct {
	vd    *VecDeque[T]
	index int
}

func (it *vecDequeReverseIterator[T]) HasNext() bool {
	return it.index >= 0
}

func (it *vecDequeReverseIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	item, _ := it.vd.Get(it.index)
	it.index--
	return item
}

// Add implements the Collection interface.
func (vd *VecDeque[T]) Add(item T) bool {
	vd.PushBack(item)
	return true
}

// Size implements the Collection interface.
func (vd *VecDeque[T]) Size() int {
	return vd.len
}

// Ensure VecDeque implements the Deque interface
var _ collections.Deque[any] = (*VecDeque[any])(nil)
