package maps

import (
	"github.com/ielm/neostd/internal/common"
	"github.com/ielm/neostd/utils/comparator"
	"github.com/ielm/neostd/utils/iterator"
)

type Map[K comparable, V any] interface {
	common.Countable
	common.Clearable
	Put(key K, value V) (V, bool)
	Get(key K) (V, bool)
	Remove(key K) (V, bool)
	ContainsKey(key K) bool
	Keys() iterator.Iterator[K]
	Values() iterator.Iterator[V]
	SetComparator(comp comparator.Comparator[K])
	Comparator() comparator.Comparator[K]
}
