// Implementation of a doubly linked list with a custom comparator

package list

import "github.com/ielm/neostd/pkg/collections"

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

func (l *LinkedList[T]) SetComparator(comp collections.Comparator[T]) {
	l.comparator = comp
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

// New method to insert in sorted order
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
