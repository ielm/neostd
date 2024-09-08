// Implementation of a doubly linked list with a custom comparator

package list

type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

type LinkedList[T any] struct {
	head       *node[T]
	tail       *node[T]
	size       int
	comparator func(T, T) bool
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

func (l *LinkedList[T]) SetComparator(comp func(T, T) bool) {
	l.comparator = comp
}

func (l *LinkedList[T]) Remove(item T) bool {
	if l.head == nil {
		return false
	}

	if l.comparator == nil {
		panic("Comparator not set for non-comparable type")
	}

	if l.comparator(l.head.value, item) {
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
		if l.comparator(current.next.value, item) {
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
