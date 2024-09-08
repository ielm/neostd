package heap

import (
	"github.com/ielm/neostd/pkg/collections"
)

// BinaryHeap is a priority queue implemented with a binary heap.
// By default, this is a max-heap. To use it as a min-heap, use the NewMinBinaryHeap function.
type BinaryHeap[T any] struct {
	data       []T
	comparator collections.Comparator[T]
}

// NewBinaryHeap creates a new BinaryHeap with the given comparator.
// This creates a max-heap by default.
func NewBinaryHeap[T any](comparator collections.Comparator[T]) *BinaryHeap[T] {
	return &BinaryHeap[T]{
		data:       make([]T, 0),
		comparator: comparator,
	}
}

// NewMinBinaryHeap creates a new BinaryHeap that functions as a min-heap.
// It uses the provided comparator but reverses the comparison.
func NewMinBinaryHeap[T any](comparator collections.Comparator[T]) *BinaryHeap[T] {
	return &BinaryHeap[T]{
		data: make([]T, 0),
		comparator: func(a, b T) int {
			return -comparator(a, b)
		},
	}
}

// Push adds an element to the heap.
// For a max-heap, this maintains the property that parent >= children.
// For a min-heap (created with NewMinBinaryHeap), this maintains the property that parent <= children.
func (h *BinaryHeap[T]) Push(item T) {
	h.data = append(h.data, item)
	h.siftUp(len(h.data) - 1)
}

// Pop removes and returns the top element from the heap.
// For a max-heap, this is the maximum element.
// For a min-heap (created with NewMinBinaryHeap), this is the minimum element.
// If the heap is empty, it returns the zero value of T and false.
func (h *BinaryHeap[T]) Pop() (T, bool) {
	if h.IsEmpty() {
		var zero T
		return zero, false
	}

	max := h.data[0]
	lastIdx := len(h.data) - 1
	h.data[0] = h.data[lastIdx]
	h.data = h.data[:lastIdx]
	if !h.IsEmpty() {
		h.siftDown(0)
	}
	return max, true
}

// Peek returns the top element without removing it.
// For a max-heap, this is the maximum element.
// For a min-heap (created with NewMinBinaryHeap), this is the minimum element.
// If the heap is empty, it returns the zero value of T and false.
func (h *BinaryHeap[T]) Peek() (T, bool) {
	if h.IsEmpty() {
		var zero T
		return zero, false
	}
	return h.data[0], true
}

// IsEmpty returns true if the heap contains no elements.
func (h *BinaryHeap[T]) IsEmpty() bool {
	return len(h.data) == 0
}

// Len returns the number of elements in the heap.
func (h *BinaryHeap[T]) Len() int {
	return len(h.data)
}

// Clear removes all elements from the heap.
func (h *BinaryHeap[T]) Clear() {
	h.data = h.data[:0]
}

// siftUp moves the element at index i up to its proper position.
func (h *BinaryHeap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.comparator(h.data[i], h.data[parent]) <= 0 {
			break
		}
		h.data[i], h.data[parent] = h.data[parent], h.data[i]
		i = parent
	}
}

// siftDown moves the element at index i down to its proper position.
func (h *BinaryHeap[T]) siftDown(i int) {
	for {
		largest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < len(h.data) && h.comparator(h.data[left], h.data[largest]) > 0 {
			largest = left
		}
		if right < len(h.data) && h.comparator(h.data[right], h.data[largest]) > 0 {
			largest = right
		}

		if largest == i {
			break
		}

		h.data[i], h.data[largest] = h.data[largest], h.data[i]
		i = largest
	}
}

// Iterator returns an iterator over the heap's elements in arbitrary order.
func (h *BinaryHeap[T]) Iterator() collections.Iterator[T] {
	return &heapIterator[T]{heap: h, index: 0}
}

type heapIterator[T any] struct {
	heap  *BinaryHeap[T]
	index int
}

func (it *heapIterator[T]) HasNext() bool {
	return it.index < len(it.heap.data)
}

func (it *heapIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	item := it.heap.data[it.index]
	it.index++
	return item
}

// IntoSortedVec returns a sorted vector of the heap's elements.
// For a max-heap, this returns the elements in descending order.
// For a min-heap (created with NewMinBinaryHeap), this returns the elements in ascending order.
func (h *BinaryHeap[T]) IntoSortedVec() []T {
	result := make([]T, len(h.data))
	for i := len(h.data) - 1; i >= 0; i-- {
		result[i], _ = h.Pop()
	}
	return result
}

// SetComparator sets a new comparator for the heap.
// This operation is O(n log n) as it requires rebuilding the heap.
func (h *BinaryHeap[T]) SetComparator(comparator collections.Comparator[T]) {
	h.comparator = comparator
	for i := len(h.data)/2 - 1; i >= 0; i-- {
		h.siftDown(i)
	}
}
