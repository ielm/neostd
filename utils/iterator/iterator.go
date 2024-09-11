package iterator

type Iterator[T any] interface {
	HasNext() bool
	Next() T
	Reset()
}

type Iterable[T any] interface {
	Iterator() Iterator[T]
	ReverseIterator() Iterator[T]
}
