package sort

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/res"
)

// CountingSort performs a counting sort on the given slice of integers.
// It returns a new sorted slice without modifying the original.
func CountingSort(slice []int) res.Result[[]int] {
	if slice == nil {
		return res.Err[[]int](errors.New(errors.ErrInvalidArgument, "input slice is nil"))
	}

	if len(slice) <= 1 {
		return res.Ok(append([]int(nil), slice...))
	}

	min, max := findMinMax(slice)
	return countingSortInRange(slice, min, max)
}

// CountingSortIterator performs a counting sort on the given iterator of integers.
// It returns a new sorted iterator without modifying the original.
func CountingSortIterator(iter collections.Iterator[int]) res.Result[collections.Iterator[int]] {
	if iter == nil {
		return res.Err[collections.Iterator[int]](errors.New(errors.ErrInvalidArgument, "input iterator is nil"))
	}

	// Collect all elements from the iterator
	var slice []int
	for iter.HasNext() {
		nextResult := iter.Next()
		if nextResult.IsSome() {
			slice = append(slice, nextResult.Unwrap())
		}
	}

	// Sort the slice
	sortedResult := CountingSort(slice)
	if sortedResult.IsErr() {
		return res.Err[collections.Iterator[int]](sortedResult.UnwrapErr())
	}

	// Return a new iterator for the sorted slice
	return res.Ok(collections.Iterator[int](NewSliceIterator(sortedResult.Unwrap())))
}

// GenericCountingSort is a generic counting sort function that can be used with any comparable type.
// It returns a new sorted slice without modifying the original.
func GenericCountingSort[T any](slice []T, keyExtractor func(T) int) res.Result[[]T] {
	if slice == nil {
		return res.Err[[]T](errors.New(errors.ErrInvalidArgument, "input slice is nil"))
	}

	if len(slice) <= 1 {
		return res.Ok(append([]T(nil), slice...))
	}

	// Extract keys and find min/max
	keys := make([]int, len(slice))
	min, max := keyExtractor(slice[0]), keyExtractor(slice[0])
	for i, item := range slice {
		key := keyExtractor(item)
		keys[i] = key
		if key < min {
			min = key
		}
		if key > max {
			max = key
		}
	}

	// Perform counting sort on keys
	sortedKeysResult := countingSortInRange(keys, min, max)
	if sortedKeysResult.IsErr() {
		return res.Err[[]T](sortedKeysResult.UnwrapErr())
	}
	sortedKeys := sortedKeysResult.Unwrap()

	// Create a map to store the original items by their keys
	itemMap := make(map[int][]T)
	for i, item := range slice {
		key := keys[i]
		itemMap[key] = append(itemMap[key], item)
	}

	// Reconstruct the sorted slice
	sortedSlice := make([]T, len(slice))
	index := 0
	for _, key := range sortedKeys {
		items := itemMap[key]
		copy(sortedSlice[index:], items)
		index += len(items)
	}

	return res.Ok(sortedSlice)
}

// findMinMax finds the minimum and maximum values in the slice.
func findMinMax(slice []int) (int, int) {
	min, max := slice[0], slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// countingSortInRange performs the actual counting sort algorithm.
func countingSortInRange(slice []int, min, max int) res.Result[[]int] {
	range_ := max - min + 1
	if range_ > 1<<32 {
		return res.Err[[]int](errors.New(errors.ErrInvalidArgument, "range too large for counting sort"))
	}

	// Count occurrences
	count := make([]int, range_)
	for _, v := range slice {
		count[v-min]++
	}

	// Calculate cumulative count
	for i := 1; i < len(count); i++ {
		count[i] += count[i-1]
	}

	// Build sorted output
	output := make([]int, len(slice))
	for i := len(slice) - 1; i >= 0; i-- {
		v := slice[i]
		count[v-min]--
		output[count[v-min]] = v
	}

	return res.Ok(output)
}

// Ensure CountingSort implements the Sort interface
var _ collections.Sort[int] = CountingSort

// CountingSortable is a wrapper struct that implements the Sortable interface for integer slices
type CountingSortable struct {
	slice []int
}

// NewCountingSortable creates a new CountingSortable instance
func NewCountingSortable(slice []int) *CountingSortable {
	return &CountingSortable{slice: slice}
}

// Sort sorts the slice in-place using Counting Sort
func (cs *CountingSortable) Sort() res.Result[collections.Sortable[int]] {
	result := CountingSort(cs.slice)
	if result.IsErr() {
		return res.Err[collections.Sortable[int]](result.UnwrapErr())
	}
	cs.slice = result.Unwrap()
	return res.Ok[collections.Sortable[int]](cs)
}

// SortWith is not applicable for Counting Sort, so it falls back to regular Sort
func (cs *CountingSortable) SortWith(less func(a, b int) bool) res.Result[collections.Sortable[int]] {
	return cs.Sort()
}

// Sorted returns a new sorted slice without modifying the original
func (cs *CountingSortable) Sorted() res.Result[collections.Sortable[int]] {
	result := CountingSort(cs.slice)
	if result.IsErr() {
		return res.Err[collections.Sortable[int]](result.UnwrapErr())
	}
	return res.Ok[collections.Sortable[int]](NewCountingSortable(result.Unwrap()))
}

// SortedWith is not applicable for Counting Sort, so it falls back to regular Sorted
func (cs *CountingSortable) SortedWith(less func(a, b int) bool) res.Result[collections.Sortable[int]] {
	return cs.Sorted()
}

// Ensure CountingSortable implements the Sortable interface
var _ collections.Sortable[int] = &CountingSortable{}

// CountingSortableIterator is a wrapper struct that implements the SortableIterator interface for integer iterators
type CountingSortableIterator struct {
	iter collections.Iterator[int]
}

// NewCountingSortableIterator creates a new CountingSortableIterator instance
func NewCountingSortableIterator(iter collections.Iterator[int]) *CountingSortableIterator {
	return &CountingSortableIterator{iter: iter}
}

// Sort sorts the iterator using Counting Sort
func (csi *CountingSortableIterator) Sort() res.Result[collections.Iterator[int]] {
	return CountingSortIterator(csi.iter)
}

// SortWith is not applicable for Counting Sort, so it falls back to regular Sort
func (csi *CountingSortableIterator) SortWith(less func(a, b int) bool) res.Result[collections.Iterator[int]] {
	return csi.Sort()
}

// Ensure CountingSortableIterator implements the SortableIterator interface
var _ collections.SortableIterator[int] = &CountingSortableIterator{}
