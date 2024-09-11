package collections

import (
	"github.com/ielm/neostd/internal/common"
	"github.com/ielm/neostd/utils/comparator"
	"github.com/ielm/neostd/utils/iterator"
)

type Collection[T any] interface {
	iterator.Iterable[T]
	common.Countable
	common.Clearable
	Add(item T) bool
	Remove(item T) bool
	Contains(item T) bool
	SetComparator(comparator.Comparator[T])
	GetComparator() comparator.Comparator[T]
}
