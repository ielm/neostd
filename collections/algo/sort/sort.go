package sort

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/res"
)

// QuickSort performs an in-place quicksort on the given slice.
// It uses the provided comparator for element comparison.
func QuickSort[T any](slice []T, comparator comp.Comparator[T]) {
	if len(slice) < 2 {
		return
	}
	quickSortRecursive(slice, 0, len(slice)-1, comparator)
}

// quickSortRecursive is the recursive helper function for QuickSort.
func quickSortRecursive[T any](slice []T, low, high int, comparator comp.Comparator[T]) {
	if low < high {
		pivotIndex := partition(slice, low, high, comparator)
		quickSortRecursive(slice, low, pivotIndex-1, comparator)
		quickSortRecursive(slice, pivotIndex+1, high, comparator)
	}
}

// partition selects a pivot and partitions the slice around it.
func partition[T any](slice []T, low, high int, comparator comp.Comparator[T]) int {
	pivot := slice[high]
	i := low - 1

	for j := low; j < high; j++ {
		if comparator(slice[j], pivot) <= 0 {
			i++
			slice[i], slice[j] = slice[j], slice[i]
		}
	}

	slice[i+1], slice[high] = slice[high], slice[i+1]
	return i + 1
}

// GenericSort is a generic sorting function that can be used with any slice type.
// It returns a new sorted slice without modifying the original.
func GenericSort[T any](slice []T, comparator comp.Comparator[T]) res.Result[[]T] {
	if slice == nil {
		return res.Err[[]T](errors.New(errors.ErrInvalidArgument, "input slice is nil"))
	}

	sortedSlice := make([]T, len(slice))
	copy(sortedSlice, slice)
	QuickSort(sortedSlice, comparator)
	return res.Ok(sortedSlice)
}

// GenericSortIterator is a generic sorting function for iterators.
// It collects elements from the iterator, sorts them, and returns a new iterator.
func GenericSortIterator[T any](iter collections.Iterator[T], comparator comp.Comparator[T]) res.Result[collections.Iterator[T]] {
	if iter == nil {
		return res.Err[collections.Iterator[T]](errors.New(errors.ErrInvalidArgument, "input iterator is nil"))
	}

	// Collect all elements from the iterator
	var slice []T
	for iter.HasNext() {
		nextResult := iter.Next()
		if nextResult.IsSome() {
			slice = append(slice, nextResult.Unwrap())
		}
	}

	// Sort the slice
	QuickSort(slice, comparator)
	// Return a new iterator for the sorted slice
	return res.Ok(collections.Iterator[T](NewSliceIterator(slice)))
}

// SliceIterator implements the Iterator interface for a slice.
type SliceIterator[T any] struct {
	slice []T
	index int
}

// NewSliceIterator creates a new SliceIterator from a slice.
func NewSliceIterator[T any](slice []T) *SliceIterator[T] {
	return &SliceIterator[T]{
		slice: slice,
		index: 0,
	}
}

// HasNext returns true if there are more elements in the iterator.
func (si *SliceIterator[T]) HasNext() bool {
	return si.index < len(si.slice)
}

// Next returns the next element in the iterator.
func (si *SliceIterator[T]) Next() res.Option[T] {
	if si.HasNext() {
		value := si.slice[si.index]
		si.index++
		return res.Some(value)
	}
	return res.None[T]()
}

// Reset resets the iterator to the beginning of the slice.
func (si *SliceIterator[T]) Reset() {
	si.index = 0
}

// Ensure SliceIterator implements the Iterator interface
var _ collections.Iterator[any] = &SliceIterator[any]{}
