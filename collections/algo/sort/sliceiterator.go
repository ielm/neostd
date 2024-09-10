package sort

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/res"
)

// sliceIterator is a simple iterator implementation for a slice
type sliceIterator[T any] struct {
	slice []T
	index int
}

func (si *sliceIterator[T]) HasNext() bool {
	return si.index < len(si.slice)
}

func (si *sliceIterator[T]) Next() res.Option[T] {
	if si.HasNext() {
		value := si.slice[si.index]
		si.index++
		return res.Some(value)
	}
	return res.None[T]()
}

// Ensure sliceIterator implements the Iterator interface
var _ collections.Iterator[int] = (*sliceIterator[int])(nil)
