// Package list provides a generic implementation of a doubly linked list.
package list

import (
	"errors"

	"github.com/ielm/neostd/pkg/collections"
)

// Node represents a single element in the linked list.
type Node[T any] struct {
	value T
	next  *Node[T]
	prev  *Node[T]
}

func (n *Node[T]) Value() T {
	return n.value
}

func (n *Node[T]) Next() *Node[T] {
	return n.next
}

func (n *Node[T]) Prev() *Node[T] {
	return n.prev
}

// LinkedList is a generic doubly linked list implementation.
type LinkedList[T any] struct {
	head       *Node[T]
	tail       *Node[T]
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

// AddFirst adds an item to the beginning of the list.
func (l *LinkedList[T]) AddFirst(item T) *Node[T] {
	newNode := &Node[T]{value: item, next: l.head}
	if l.head != nil {
		l.head.prev = newNode
	} else {
		l.tail = newNode
	}
	l.head = newNode
	l.size++
	return newNode
}

// AddLast adds an item to the end of the list.
func (l *LinkedList[T]) AddLast(item T) *Node[T] {
	newNode := &Node[T]{value: item, prev: l.tail}
	if l.tail != nil {
		l.tail.next = newNode
	} else {
		l.head = newNode
	}
	l.tail = newNode
	l.size++
	return newNode
}

// AddAfter adds a new item after the specified node.
func (l *LinkedList[T]) AddAfter(node *Node[T], item T) *Node[T] {
	if node == nil {
		return l.AddFirst(item)
	}

	newNode := &Node[T]{value: item, prev: node, next: node.next}
	if node.next != nil {
		node.next.prev = newNode
	} else {
		l.tail = newNode
	}
	node.next = newNode
	l.size++
	return newNode
}

// MoveToFront moves the first occurrence of the specified item to the front of the list.
func (l *LinkedList[T]) MoveToFront(item T) bool {
	if l.comparator == nil {
		panic("comparator not set for non-comparable type")
	}

	current := l.head
	for current != nil {
		if l.comparator(current.value, item) == 0 {
			l.MoveNodeToFront(current)
			return true
		}
		current = current.next
	}
	return false
}

// MoveNodeToFront moves the specified node to the front of the list.
func (l *LinkedList[T]) MoveNodeToFront(node *Node[T]) {
	if node == nil || node == l.head {
		return
	}

	// Remove the node from its current position
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev // Update tail if moving the last node
	}

	// Move the node to the front
	node.prev = nil
	node.next = l.head
	l.head.prev = node
	l.head = node
}

// MoveToEnd moves the first occurrence of the specified item to the end of the list.
func (l *LinkedList[T]) MoveToEnd(item T) bool {
	if l.comparator == nil {
		panic("comparator not set for non-comparable type")
	}

	current := l.head
	for current != nil {
		if l.comparator(current.value, item) == 0 {
			l.MoveNodeToEnd(current)
			return true
		}
		current = current.next
	}
	return false
}

// MoveNodeToEnd moves the specified node to the end of the list.
func (l *LinkedList[T]) MoveNodeToEnd(node *Node[T]) {
	if node == nil || node == l.tail {
		return
	}

	// Remove the node from its current position
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}

	// Move the node to the end
	node.next = nil
	node.prev = l.tail
	l.tail.next = node
	l.tail = node
}

// RemoveNode removes a specific node from the list.
func (l *LinkedList[T]) RemoveNode(n *Node[T]) {
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

// Remove removes the first occurrence of the specified item from the list.
func (l *LinkedList[T]) Remove(item T) bool {
	current := l.head
	for current != nil {
		if l.comparator(current.value, item) == 0 {
			l.RemoveNode(current)
			return true
		}
		current = current.next
	}
	return false
}

// RemoveLast removes and returns the last item in the list.
func (l *LinkedList[T]) RemoveLast() (T, bool) {
	if l.IsEmpty() {
		var zero T
		return zero, false
	}
	value := l.tail.value
	l.RemoveNode(l.tail)
	return value, true
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
func (l *LinkedList[T]) getNode(index int) *Node[T] {
	if index < l.size/2 {
		return l.traverseForward(index)
	}
	return l.traverseBackward(index)
}

// RemoveFirst removes and returns the first item in the list.
func (l *LinkedList[T]) RemoveFirst() (T, error) {
	if l.IsEmpty() {
		var zero T
		return zero, errors.New("list is empty")
	}
	value := l.head.value
	l.RemoveNode(l.head)
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
	l.RemoveNode(node)
	return value, nil
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
func (l *LinkedList[T]) traverseForward(index int) *Node[T] {
	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	return current
}

// traverseBackward traverses the list from the tail to the specified index.
func (l *LinkedList[T]) traverseBackward(index int) *Node[T] {
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
	current *Node[T]
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
	current *Node[T]
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

// First returns a pointer to the first node in the list.
func (l *LinkedList[T]) First() *Node[T] {
	return l.head
}

// Last returns a pointer to the last node in the list.
func (l *LinkedList[T]) Last() *Node[T] {
	if l.IsEmpty() {
		return nil
	}
	return l.tail
}
