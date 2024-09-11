package tree

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/hash"
	"github.com/ielm/neostd/res"
)

// Tree is the interface that all tree implementations should follow
type Tree[K comparable, V any] interface {
	collections.Map[K, V]
	Root() *Node[K, V]
	Insert(key K, value V) error
	Delete(key K) error
	Search(key K) (*Node[K, V], bool)
	Traverse(order TraversalOrder) []collections.Pair[K, V]
}

// Node represents a node in the tree
type Node[K any, V any] struct {
	Key      K
	Value    V
	Children []*Node[K, V]
}

// TraversalOrder defines the order of tree traversal
type TraversalOrder int

const (
	PreOrder TraversalOrder = iota
	InOrder
	PostOrder
	LevelOrder
)

// BaseTree is the common implementation for all trees
type BaseTree[K any, V any] struct {
	root       *Node[K, V]
	size       int
	comparator comp.Comparator[K]
	hasher     hash.Hasher
}

// NewBaseTree creates a new base tree
func NewBaseTree[K any, V any](comparator comp.Comparator[K], hasher hash.Hasher) *BaseTree[K, V] {
	return &BaseTree[K, V]{
		comparator: comparator,
		hasher:     hasher,
	}
}

// Root returns the root node of the tree
func (t *BaseTree[K, V]) Root() *Node[K, V] {
	return t.root
}

// Size returns the number of nodes in the tree
func (t *BaseTree[K, V]) Size() int {
	return t.size
}

// Clear removes all nodes from the tree
func (t *BaseTree[K, V]) Clear() {
	t.root = nil
	t.size = 0
}

// IsEmpty returns true if the tree has no nodes
func (t *BaseTree[K, V]) IsEmpty() bool {
	return t.size == 0
}

// SetComparator sets the comparator for the tree
func (t *BaseTree[K, V]) SetComparator(comp comp.Comparator[K]) {
	t.comparator = comp
}

// SetHasher sets the hasher for the tree
func (t *BaseTree[K, V]) SetHasher(h hash.Hasher) {
	t.hasher = h
}

// Comparator returns the comparator for the tree
func (t *BaseTree[K, V]) Comparator() comp.Comparator[K] {
	return t.comparator
}

// Hasher returns the hasher for the tree
func (t *BaseTree[K, V]) Hasher() hash.Hasher {
	return t.hasher
}

// Iterator returns an iterator over the tree's nodes
func (t *BaseTree[K, V]) Iterator() collections.Iterator[collections.Pair[K, V]] {
	return &treeIterator[K, V]{
		tree:  t,
		stack: []*Node[K, V]{t.root},
	}
}

// ReverseIterator returns a reverse iterator over the tree's nodes
func (t *BaseTree[K, V]) ReverseIterator() collections.Iterator[collections.Pair[K, V]] {
	// Implementation depends on the specific tree type
	panic("ReverseIterator not implemented for base tree")
}

type treeIterator[K any, V any] struct {
	tree  *BaseTree[K, V]
	stack []*Node[K, V]
}

func (it *treeIterator[K, V]) HasNext() bool {
	return len(it.stack) > 0
}

func (it *treeIterator[K, V]) Next() res.Option[collections.Pair[K, V]] {
	if !it.HasNext() {
		return res.None[collections.Pair[K, V]]()
	}
	node := it.stack[len(it.stack)-1]
	it.stack = it.stack[:len(it.stack)-1]
	for i := len(node.Children) - 1; i >= 0; i-- {
		it.stack = append(it.stack, node.Children[i])
	}
	return res.Some(collections.Pair[K, V]{Key: node.Key, Value: node.Value})
}

// Ensure baseTree implements the Collection interface
// var _ collections.Collection[int] = (*baseTree[int])(nil)
