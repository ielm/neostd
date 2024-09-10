package tree

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/hash"
	"github.com/ielm/neostd/res"
)

// Tree is the interface that all tree implementations should follow
type Tree[T any] interface {
	collections.Collection[T]
	Root() *Node[T]
	Insert(value T) error
	Delete(value T) error
	Search(value T) (*Node[T], bool)
	Traverse(order TraversalOrder) []T
}

// Node represents a node in the tree
type Node[T any] struct {
	Value    T
	Children []*Node[T]
}

// TraversalOrder defines the order of tree traversal
type TraversalOrder int

const (
	PreOrder TraversalOrder = iota
	InOrder
	PostOrder
	LevelOrder
)

// baseTree is the common implementation for all trees
type baseTree[T any] struct {
	root       *Node[T]
	size       int
	comparator comp.Comparator[T]
	hasher     hash.Hasher
}

// newBaseTree creates a new base tree
func newBaseTree[T any](comparator comp.Comparator[T], hasher hash.Hasher) *baseTree[T] {
	return &baseTree[T]{
		comparator: comparator,
		hasher:     hasher,
	}
}

// Root returns the root node of the tree
func (t *baseTree[T]) Root() *Node[T] {
	return t.root
}

// Size returns the number of nodes in the tree
func (t *baseTree[T]) Size() int {
	return t.size
}

// Clear removes all nodes from the tree
func (t *baseTree[T]) Clear() {
	t.root = nil
	t.size = 0
}

// IsEmpty returns true if the tree has no nodes
func (t *baseTree[T]) IsEmpty() bool {
	return t.size == 0
}

// SetComparator sets the comparator for the tree
func (t *baseTree[T]) SetComparator(comp comp.Comparator[T]) {
	t.comparator = comp
}

// SetHasher sets the hasher for the tree
func (t *baseTree[T]) SetHasher(h hash.Hasher) {
	t.hasher = h
}

// Iterator returns an iterator over the tree's nodes
func (t *baseTree[T]) Iterator() collections.Iterator[T] {
	return &treeIterator[T]{
		tree:  t,
		stack: []*Node[T]{t.root},
	}
}

// ReverseIterator returns a reverse iterator over the tree's nodes
func (t *baseTree[T]) ReverseIterator() collections.Iterator[T] {
	// Implementation depends on the specific tree type
	panic("ReverseIterator not implemented for base tree")
}

type treeIterator[T any] struct {
	tree  *baseTree[T]
	stack []*Node[T]
}

func (it *treeIterator[T]) HasNext() bool {
	return len(it.stack) > 0
}

func (it *treeIterator[T]) Next() res.Option[T] {
	if !it.HasNext() {
		return res.None[T]()
	}
	node := it.stack[len(it.stack)-1]
	it.stack = it.stack[:len(it.stack)-1]
	for i := len(node.Children) - 1; i >= 0; i-- {
		it.stack = append(it.stack, node.Children[i])
	}
	return res.Some(node.Value)
}

// Ensure baseTree implements the Collection interface
// var _ collections.Collection[int] = (*baseTree[int])(nil)
