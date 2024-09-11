package queue

import (
	"github.com/ielm/neostd/collections/list"
	"github.com/ielm/neostd/utils/result"
)

type Deque[T any] interface {
	list.List[T]
	PushFront(item T)
	PushBack(item T)
	PopFront() result.Option[T]
	PopBack() result.Option[T]
	Front() result.Option[T]
	Back() result.Option[T]
}
