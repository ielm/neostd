package sort

import (
	"github.com/ielm/neostd/utils/iterator"
	"github.com/ielm/neostd/utils/result"
)

type Sort[T any] interface {
	Sort([]T) result.Result[[]T]
	SortIterator(iterator.Iterator[T]) result.Result[iterator.Iterator[T]]
}

type Sortable[T any] interface {
	Sort() result.Result[Sortable[T]]
	SortWith(less func(a, b T) bool) result.Result[Sortable[T]]
	Sorted() result.Result[Sortable[T]]
	SortedWith(less func(a, b T) bool) result.Result[Sortable[T]]
}
