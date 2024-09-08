package collections

// Iterable represents a collection that can be iterated over
type Iterable[T any] interface {
	Iterator() Iterator[T]
	ReverseIterator() Iterator[T] // New method for reverse iteration
}

// Iterator represents an iterator over a collection
type Iterator[T any] interface {
	HasNext() bool
	Next() T
}

type Pair[K any, V any] struct {
	Key   K
	Value V
}

type Collection[T any] interface {
	Iterable[T]
	Add(item T) bool
	Remove(item T) bool
	Contains(item T) bool
	Size() int
	Clear()
	IsEmpty() bool
	SetComparator(comp func(T, T) bool)
}

// List represents an ordered collection
type List[T any] interface {
	Collection[T]
	Get(index int) (T, error)
	Set(index int, item T) error
	IndexOf(item T) int
}

// Set represents a collection that contains no duplicate elements
type Set[T any] interface {
	Collection[T]
}

// Map represents a collection of key-value pairs
type Map[K comparable, V any] interface {
	Put(key K, value V) (V, bool)
	Get(key K) (V, bool)
	Remove(key K) (V, bool)
	ContainsKey(key K) bool
	Keys() []K
	Values() []V
	SetComparator(comp Comparator[K])
}
