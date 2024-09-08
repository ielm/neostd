// Implementation of a doubly linked list with a custom comparator

package list

import (
	"errors"

	"github.com/ielm/neostd/pkg/collections"
)

type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

type LinkedList[T any] struct {
	head       *node[T]
	tail       *node[T]
	size       int
	comparator collections.Comparator[T]
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

func (l *LinkedList[T]) SetComparator(comp collections.Comparator[T]) {
	l.comparator = comp
}

func (l *LinkedList[T]) Add(item T) bool {
	newNode := &node[T]{value: item}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.tail.next = newNode
		newNode.prev = l.tail
		l.tail = newNode
	}
	l.size++

	return true
}

// AddFirst adds an item to the beginning of the list
func (l *LinkedList[T]) AddFirst(item T) {
	newNode := &node[T]{value: item}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		newNode.next = l.head
		l.head.prev = newNode
		l.head = newNode
	}
	l.size++
}

// AddLast adds an item to the end of the list
func (l *LinkedList[T]) AddLast(item T) {
	l.Add(item) // This is the same as Add
}

func (l *LinkedList[T]) Set(index int, item T) error {
	if index < 0 || index >= l.size {
		return errors.New("index out of bounds")
	}

	if index < l.size/2 {
		return l.setForward(index, item)
	} else {
		return l.setBackward(index, item)
	}
}

func (l *LinkedList[T]) setForward(index int, item T) error {
	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	current.value = item
	return nil
}

func (l *LinkedList[T]) setBackward(index int, item T) error {
	current := l.tail
	for i := l.size - 1; i > index; i-- {
		current = current.prev
	}
	current.value = item
	return nil
}

func (l *LinkedList[T]) InsertSorted(item T) {
	if l.comparator == nil {
		panic("Comparator not set for non-comparable type")
	}

	newNode := &node[T]{value: item}

	if l.head == nil || l.comparator(item, l.head.value) <= 0 {
		newNode.next = l.head
		if l.head != nil {
			l.head.prev = newNode
		}
		l.head = newNode
		if l.tail == nil {
			l.tail = newNode
		}
	} else {
		current := l.head
		for current.next != nil && l.comparator(item, current.next.value) > 0 {
			current = current.next
		}
		newNode.next = current.next
		newNode.prev = current
		if current.next != nil {
			current.next.prev = newNode
		} else {
			l.tail = newNode
		}
		current.next = newNode
	}

	l.size++
}

func (l *LinkedList[T]) Remove(item T) bool {
	if l.head == nil {
		return false
	}

	if l.comparator == nil {
		panic("Comparator not set for non-comparable type")
	}

	if l.comparator(l.head.value, item) == 0 {
		l.head = l.head.next
		if l.head == nil {
			l.tail = nil
		} else {
			l.head.prev = nil
		}
		l.size--
		return true
	}

	current := l.head
	for current.next != nil {
		if l.comparator(current.next.value, item) == 0 {
			current.next = current.next.next
			if current.next == nil {
				l.tail = current
			} else {
				current.next.prev = current
			}
			l.size--
			return true
		}
		current = current.next
	}

	return false
}

// RemoveFirst removes and returns the first item in the list
func (l *LinkedList[T]) RemoveFirst() (T, error) {
	if l.head == nil {
		var zero T
		return zero, errors.New("list is empty")
	}
	value := l.head.value
	l.head = l.head.next
	if l.head == nil {
		l.tail = nil
	} else {
		l.head.prev = nil
	}
	l.size--
	return value, nil
}

// RemoveLast removes and returns the last item in the list
func (l *LinkedList[T]) RemoveLast() (T, error) {
	if l.tail == nil {
		var zero T
		return zero, errors.New("list is empty")
	}
	value := l.tail.value
	l.tail = l.tail.prev
	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.next = nil
	}
	l.size--
	return value, nil
}

func (l *LinkedList[T]) RemoveAt(index int) (T, error) {
	if index < 0 || index >= l.size {
		var zero T
		return zero, errors.New("index out of bounds")
	}

	if index < l.size/2 {
		return l.removeAtForward(index)
	} else {
		return l.removeAtBackward(index)
	}
}

func (l *LinkedList[T]) removeAtForward(index int) (T, error) {
	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	value := current.value
	current.next.prev = current.prev

	if current.prev != nil {
		current.prev.next = current.next
	} else {
		l.head = current.next
	}

	l.size--
	return value, nil
}

func (l *LinkedList[T]) removeAtBackward(index int) (T, error) {
	current := l.tail
	for i := l.size - 1; i > index; i-- {
		current = current.prev
	}
	value := current.value
	current.prev.next = current.next

	if current.next != nil {
		current.next.prev = current.prev
	} else {
		l.tail = current.prev
	}

	l.size--
	return value, nil
}

func (l *LinkedList[T]) Contains(item T) bool {
	if l.head == nil {
		return false
	}

	if l.comparator == nil {
		panic("Comparator not set for non-comparable type")
	}

	current := l.head
	for current != nil {
		if l.comparator(current.value, item) == 0 {
			return true
		}
		current = current.next
	}

	return false
}

func (l *LinkedList[T]) Size() int {
	return l.size
}

func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}

func (l *LinkedList[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

func (l *LinkedList[T]) Get(index int) (T, error) {
	if index < 0 || index >= l.size {
		var zero T
		return zero, errors.New("index out of bounds")
	}

	if index < l.size/2 {
		return l.traverseForward(index).value, nil

	} else {
		return l.traverseBackward(index).value, nil
	}
}

func (l *LinkedList[T]) traverseForward(index int) *node[T] {
	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	return current
}

func (l *LinkedList[T]) traverseBackward(index int) *node[T] {
	current := l.tail
	for i := l.size - 1; i > index; i-- {
		current = current.prev
	}
	return current
}

func (l *LinkedList[T]) IndexOf(item T) int {
	if l.comparator == nil {
		panic("Comparator not set for non-comparable type")
	}

	current := l.head
	for i := 0; i < l.size; i++ {
		if l.comparator(current.value, item) == 0 {
			return i
		}
		current = current.next
	}
	return -1
}
