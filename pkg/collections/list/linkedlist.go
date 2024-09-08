// Package list provides a generic implementation of a doubly linked list.
package list

import (
	"errors"

	"github.com/ielm/neostd/pkg/collections"
)

// node represents a single element in the linked list.
type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

// LinkedList is a generic doubly linked list implementation.
type LinkedList[T any] struct {
	head       *node[T]
	tail       *node[T]
	size       int
	comparator collections.Comparator[T]
}

// NewLinkedList creates and returns a new empty LinkedList.
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// SetComparator sets the comparator function for the list.
func (l *LinkedList[T]) SetComparator(comp collections.Comparator[T]) {
	l.comparator = comp
}

// Add appends an item to the end of the list.
func (l *LinkedList[T]) Add(item T) bool {
	l.addLast(item)
	return true
}

// AddFirst adds an item to the beginning of the list.
func (l *LinkedList[T]) AddFirst(item T) {
	newNode := &node[T]{value: item, next: l.head}
	if l.head != nil {
		l.head.prev = newNode
	} else {
		l.tail = newNode
	}
	l.head = newNode
	l.size++
}

// AddLast adds an item to the end of the list.
func (l *LinkedList[T]) AddLast(item T) {
	l.addLast(item)
}

// addLast is a helper method to add an item to the end of the list.
func (l *LinkedList[T]) addLast(item T) {
	newNode := &node[T]{value: item, prev: l.tail}
	if l.tail != nil {
		l.tail.next = newNode
	} else {
		l.head = newNode
	}
	l.tail = newNode
	l.size++
}

// Set updates the item at the specified index.
func (l *LinkedList[T]) Set(index int, item T) error {
	if index < 0 || index >= l.size {
		return errors.New("index out of bounds")
	}

	node := l.getNode(index)
	node.value = item
	return nil
}

// InsertSorted inserts an item into the list maintaining sorted order.
func (l *LinkedList[T]) InsertSorted(item T) {
	if l.comparator == nil {
		panic("comparator not set for non-comparable type")
	}

	newNode := &node[T]{value: item}

	if l.head == nil || l.comparator(item, l.head.value) <= 0 {
		l.AddFirst(item)
		return
	}

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
	l.size++
}

// Get returns the item at the specified index.
func (l *LinkedList[T]) Get(index int) (T, error) {
	if index < 0 || index >= l.size {
		var zero T
		return zero, errors.New("index out of bounds")
	}
	return l.getNode(index).value, nil
}

// getNode returns the node at the specified index.
func (l *LinkedList[T]) getNode(index int) *node[T] {
	if index < l.size/2 {
		return l.traverseForward(index)
	}
	return l.traverseBackward(index)
}

// Remove removes the first occurrence of the specified item from the list.
func (l *LinkedList[T]) Remove(item T) bool {
	if l.comparator == nil {
		panic("comparator not set for non-comparable type")
	}

	current := l.head
	for current != nil {
		if l.comparator(current.value, item) == 0 {
			l.removeNode(current)
			return true
		}
		current = current.next
	}
	return false
}

// RemoveFirst removes and returns the first item in the list.
func (l *LinkedList[T]) RemoveFirst() (T, error) {
	if l.IsEmpty() {
		var zero T
		return zero, errors.New("list is empty")
	}
	value := l.head.value
	l.removeNode(l.head)
	return value, nil
}

// RemoveLast removes and returns the last item in the list.
func (l *LinkedList[T]) RemoveLast() (T, error) {
	if l.IsEmpty() {
		var zero T
		return zero, errors.New("list is empty")
	}
	value := l.tail.value
	l.removeNode(l.tail)
	return value, nil
}

// RemoveAt removes and returns the item at the specified index.
func (l *LinkedList[T]) RemoveAt(index int) (T, error) {
	if index < 0 || index >= l.size {
		var zero T
		return zero, errors.New("index out of bounds")
	}

	node := l.getNode(index)
	value := node.value
	l.removeNode(node)
	return value, nil
}

// removeNode is a helper method to remove a node from the list.
func (l *LinkedList[T]) removeNode(n *node[T]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}
	l.size--
}

// Contains checks if the list contains the specified item.
func (l *LinkedList[T]) Contains(item T) bool {
	if l.comparator == nil {
		panic("comparator not set for non-comparable type")
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

// Size returns the number of elements in the list.
func (l *LinkedList[T]) Size() int {
	return l.size
}

// IsEmpty returns true if the list is empty.
func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}

// Clear removes all elements from the list.
func (l *LinkedList[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

// traverseForward traverses the list from the head to the specified index.
func (l *LinkedList[T]) traverseForward(index int) *node[T] {
	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	return current
}

// traverseBackward traverses the list from the tail to the specified index.
func (l *LinkedList[T]) traverseBackward(index int) *node[T] {
	current := l.tail
	for i := l.size - 1; i > index; i-- {
		current = current.prev
	}
	return current
}

// IndexOf returns the index of the first occurrence of the specified item.
func (l *LinkedList[T]) IndexOf(item T) int {
	if l.comparator == nil {
		panic("comparator not set for non-comparable type")
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

// Iterator returns an iterator for the list.
func (l *LinkedList[T]) Iterator() collections.Iterator[T] {
	return &linkedListIterator[T]{current: l.head}
}

// ReverseIterator returns a reverse iterator for the list.
func (l *LinkedList[T]) ReverseIterator() collections.Iterator[T] {
	return &linkedListReverseIterator[T]{current: l.tail}
}

// linkedListIterator implements the Iterator interface for LinkedList.
type linkedListIterator[T any] struct {
	current *node[T]
}

func (it *linkedListIterator[T]) HasNext() bool {
	return it.current != nil
}

func (it *linkedListIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	value := it.current.value
	it.current = it.current.next
	return value
}

// linkedListReverseIterator implements the Iterator interface for reverse iteration.
type linkedListReverseIterator[T any] struct {
	current *node[T]
}

func (it *linkedListReverseIterator[T]) HasNext() bool {
	return it.current != nil
}

func (it *linkedListReverseIterator[T]) Next() T {
	if !it.HasNext() {
		panic("no more elements")
	}
	value := it.current.value
	it.current = it.current.prev
	return value
}
