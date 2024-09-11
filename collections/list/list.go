package list

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/utils/result"
)

type List[T any] interface {
	collections.Collection[T]
	Get(index int) result.Result[T]
	Set(index int, item T) result.Result[T]
	IndexOf(item T) result.Option[int]
}

type Vector[T any] interface {
	List[T]
	Push(item T)
	Pop() result.Option[T]
	Cap() int
	Grow(newCap int)
}
