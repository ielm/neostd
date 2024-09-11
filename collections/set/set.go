package set

import "github.com/ielm/neostd/collections"

type Set[T any] interface {
	collections.Collection[T]
}

type SortedSet[T any] interface {
	Set[T]
	First() (T, bool)
	Last() (T, bool)
	Floor(item T) (T, bool)
	Ceiling(item T) (T, bool)
}
