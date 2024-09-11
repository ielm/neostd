package collections

import (
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/res"
)

// Iterable represents a collection that can be iterated over
type Iterable[T any] interface {
	Iterator() Iterator[T]
	ReverseIterator() Iterator[T] // New method for reverse iteration
}

// Iterator represents an iterator over a collection
type Iterator[T any] interface {
	HasNext() bool
	Next() res.Option[T]
}

type Pair[K any, V any] struct {
	Key   K
	Value V
}

type Countable interface {
	Size() int
	IsEmpty() bool
}

type Clearable interface {
	Clear()
}

type Collection[T any] interface {
	Iterable[T]
	Countable
	Clearable
	Add(item T) bool
	Remove(item T) bool
	Contains(item T) bool
	SetComparator(comp comp.Comparator[T])
	Comparator() comp.Comparator[T]
}

// List represents an ordered collection
type List[T any] interface {
	Collection[T]
	Get(index int) res.Result[T]
	Set(index int, item T) res.Result[T]
	IndexOf(item T) res.Option[int]
}

// Vector represents a resizable array
type Vector[T any] interface {
	List[T]
	Push(item T)
	Pop() (T, bool)
	Cap() int
	Grow(newCap int)
}

type Deque[T any] interface {
	List[T]
	PushFront(item T)
	PushBack(item T)
	PopFront() (T, bool)
	PopBack() (T, bool)
	Front() (T, bool)
	Back() (T, bool)
}

// Set represents a collection that contains no duplicate elements
type Set[T any] interface {
	Collection[T]
}

// Map represents a collection of key-value pairs
//
//	type Map[K comparable, V any] interface {
//		Put(key K, value V) (V, bool)
//		Get(key K) (V, bool)
//		Remove(key K) (V, bool)
//		ContainsKey(key K) bool
//		Keys() []K
//		Values() []V
//		SetComparator(comp comp.Comparator[K])
//	}

type Map[K comparable, V any] interface {
	Countable
	Clearable
	Put(key K, value V) (V, bool)
	Get(key K) (V, bool)
	Remove(key K) (V, bool)
	ContainsKey(key K) bool
	Keys() []K
	Values() []V
	SetComparator(comp comp.Comparator[K])
	Comparator() comp.Comparator[K]
}

// ProbabilisticSet represents a probabilistic set data structure
type ProbabilisticSet[T any] interface {
	Countable
	Clearable
	Add(item T) bool
	Contains(item T) bool
}

// SortedSet represents an ordered set with additional operations
type SortedSet[T any] interface {
	Set[T]
	Get(item T) (T, bool)
}

// Sort defines the interface for sorting algorithms that operate on slices.
type Sort[T any] func([]T) res.Result[[]T]

// SortIterator defines the interface for sorting algorithms that operate on iterators.
type SortIterator[T any] func(Iterator[T]) Iterator[T]

// Sortable is an interface that defines methods for sorting a collection.
type Sortable[T any] interface {
	// Sort sorts the collection in-place and returns the sorted collection.
	Sort() res.Result[Sortable[T]]

	// SortWith sorts the collection in-place using the provided comparison function.
	SortWith(less func(a, b T) bool) res.Result[Sortable[T]]

	// Sorted returns a new sorted collection without modifying the original.
	Sorted() res.Result[Sortable[T]]

	// SortedWith returns a new sorted collection using the provided comparison function,
	// without modifying the original.
	SortedWith(less func(a, b T) bool) res.Result[Sortable[T]]
}

// SortableIterator is an interface that defines methods for sorting iterators.
type SortableIterator[T any] interface {
	// Sort sorts the iterator and returns a new sorted iterator.
	Sort() res.Result[Iterator[T]]

	// SortWith sorts the iterator using the provided comparison function
	// and returns a new sorted iterator.
	SortWith(less func(a, b T) bool) res.Result[Iterator[T]]
}
