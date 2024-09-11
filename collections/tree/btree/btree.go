package btree

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/collections/tree"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/hash"
)

const (
	defaultDegree = 64
	minDegree     = 2
)

// BTree represents a B-tree data structure.
// It implements both the Tree and Map interfaces.
type BTree[K any, V any] struct {
	root       *node[K, V]
	degree     int
	size       int
	comparator comp.Comparator[K]
	hasher     hash.Hasher
}

// node represents a single node in the BTree.
type node[K any, V any] struct {
	keys     []K
	values   []V
	children []*node[K, V]
	leaf     bool
}

// New creates a new BTree with the specified degree, comparator, and hasher.
// If the degree is less than minDegree, it defaults to defaultDegree.
func New[K any, V any](degree int, comparator comp.Comparator[K], hasher hash.Hasher) *BTree[K, V] {
	if degree < minDegree {
		degree = defaultDegree
	}
	return &BTree[K, V]{
		degree:     degree,
		comparator: comparator,
		hasher:     hasher,
	}
}

// SetComparator sets the comparator for the BTree.
func (t *BTree[K, V]) SetComparator(comparator comp.Comparator[K]) {
	t.comparator = comparator
}

// Comparator returns the comparator for the BTree.
func (t *BTree[K, V]) Comparator() comp.Comparator[K] {
	return t.comparator
}

// Root returns the root node of the tree.
func (t *BTree[K, V]) Root() *tree.Node[K, V] {
	if t.root == nil {
		return nil
	}
	return &tree.Node[K, V]{
		Key:   t.root.keys[0],
		Value: t.root.values[0],
	}
}

// Insert inserts a key-value pair into the BTree.
func (t *BTree[K, V]) Insert(key K, value V) error {
	if t.root == nil {
		t.root = t.createNode(true)
		t.root.keys = append(t.root.keys, key)
		t.root.values = append(t.root.values, value)
		t.size++
		return nil
	}

	if len(t.root.keys) == 2*t.degree-1 {
		newRoot := t.createNode(false)
		newRoot.children = append(newRoot.children, t.root)
		t.splitChild(newRoot, 0)
		t.root = newRoot
	}

	t.insertNonFull(t.root, key, value)
	t.size++
	return nil
}

// Delete removes a key and its associated value from the BTree.
func (t *BTree[K, V]) Delete(key K) error {
	if t.root == nil {
		return errors.New(errors.ErrNotFound, "key not found")
	}

	found, err := t.delete(t.root, key)
	if !found {
		return err
	}

	if len(t.root.keys) == 0 && !t.root.leaf {
		t.root = t.root.children[0]
	}

	t.size--
	return nil
}

// Search searches for a key in the BTree.
func (t *BTree[K, V]) Search(key K) (*tree.Node[K, V], bool) {
	if t.root == nil {
		return nil, false
	}

	node, index := t.search(t.root, key)
	if index < len(node.keys) && t.comparator(node.keys[index], key) == 0 {
		return &tree.Node[K, V]{
			Key:   node.keys[index],
			Value: node.values[index],
		}, true
	}

	return nil, false
}

// Size returns the number of key-value pairs in the BTree.
func (t *BTree[K, V]) Size() int {
	return t.size
}

// Clear removes all elements from the BTree.
func (t *BTree[K, V]) Clear() {
	t.root = nil
	t.size = 0
}

// IsEmpty returns true if the BTree is empty.
func (t *BTree[K, V]) IsEmpty() bool {
	return t.size == 0
}

// Traverse traverses the BTree in the specified order.
func (t *BTree[K, V]) Traverse(order tree.TraversalOrder) []collections.Pair[K, V] {
	result := make([]collections.Pair[K, V], 0, t.size)
	if t.root == nil {
		return result
	}

	switch order {
	case tree.InOrder:
		t.inOrderTraversal(t.root, &result)
	case tree.PreOrder:
		t.preOrderTraversal(t.root, &result)
	case tree.PostOrder:
		t.postOrderTraversal(t.root, &result)
	case tree.LevelOrder:
		t.levelOrderTraversal(t.root, &result)
	}

	return result
}

// Helper methods

// createNode creates a new node for the BTree.
func (t *BTree[K, V]) createNode(leaf bool) *node[K, V] {
	return &node[K, V]{
		keys:     make([]K, 0, 2*t.degree-1),
		values:   make([]V, 0, 2*t.degree-1),
		children: make([]*node[K, V], 0, 2*t.degree),
		leaf:     leaf,
	}
}

// splitChild splits a full child node during insertion.
func (t *BTree[K, V]) splitChild(parent *node[K, V], index int) {
	child := parent.children[index]
	newChild := t.createNode(child.leaf)

	parent.keys = append(parent.keys, child.keys[t.degree-1])
	parent.values = append(parent.values, child.values[t.degree-1])
	parent.children = append(parent.children, nil)

	copy(parent.children[index+2:], parent.children[index+1:])
	parent.children[index+1] = newChild

	newChild.keys = append(newChild.keys, child.keys[t.degree:]...)
	newChild.values = append(newChild.values, child.values[t.degree:]...)
	child.keys = child.keys[:t.degree-1]
	child.values = child.values[:t.degree-1]

	if !child.leaf {
		newChild.children = append(newChild.children, child.children[t.degree:]...)
		child.children = child.children[:t.degree]
	}
}

// insertNonFull inserts a key-value pair into a non-full node.
func (t *BTree[K, V]) insertNonFull(n *node[K, V], key K, value V) {
	i := len(n.keys) - 1

	if n.leaf {
		n.keys = append(n.keys, key)
		n.values = append(n.values, value)
		for i >= 0 && t.comparator(key, n.keys[i]) < 0 {
			n.keys[i+1] = n.keys[i]
			n.values[i+1] = n.values[i]
			i--
		}
		n.keys[i+1] = key
		n.values[i+1] = value
	} else {
		for i >= 0 && t.comparator(key, n.keys[i]) < 0 {
			i--
		}
		i++
		if len(n.children[i].keys) == 2*t.degree-1 {
			t.splitChild(n, i)
			if t.comparator(key, n.keys[i]) > 0 {
				i++
			}
		}
		t.insertNonFull(n.children[i], key, value)
	}
}

// delete removes a key and its associated value from the BTree.
func (t *BTree[K, V]) delete(n *node[K, V], key K) (bool, error) {
	i := 0
	for i < len(n.keys) && t.comparator(key, n.keys[i]) > 0 {
		i++
	}

	if i < len(n.keys) && t.comparator(key, n.keys[i]) == 0 {
		if n.leaf {
			t.deleteFromLeaf(n, i)
		} else {
			t.deleteFromInternalNode(n, i)
		}
	} else if !n.leaf {
		t.deleteFromNonLeaf(n, i)
	} else {
		return false, errors.New(errors.ErrNotFound, "key not found")
	}

	return true, nil
}

// deleteFromLeaf removes a key-value pair from a leaf node.
func (t *BTree[K, V]) deleteFromLeaf(n *node[K, V], index int) {
	copy(n.keys[index:], n.keys[index+1:])
	copy(n.values[index:], n.values[index+1:])
	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]
}

// deleteFromInternalNode removes a key-value pair from an internal node.
func (t *BTree[K, V]) deleteFromInternalNode(n *node[K, V], index int) {
	key := n.keys[index]

	if len(n.children[index].keys) >= t.degree {
		predecessor := t.getPredecessor(n, index)
		n.keys[index] = predecessor
		n.values[index] = n.children[index].values[len(n.children[index].keys)-1]
		t.delete(n.children[index], predecessor)
	} else if len(n.children[index+1].keys) >= t.degree {
		successor := t.getSuccessor(n, index)
		n.keys[index] = successor
		n.values[index] = n.children[index+1].values[0]
		t.delete(n.children[index+1], successor)
	} else {
		t.mergeChildren(n, index)
		t.delete(n.children[index], key)
	}
}

// deleteFromNonLeaf removes a key-value pair from a non-leaf node.
func (t *BTree[K, V]) deleteFromNonLeaf(n *node[K, V], index int) {
	key := n.keys[index]

	if len(n.children[index].keys) >= t.degree {
		predecessor := t.getPredecessor(n, index)
		n.keys[index] = predecessor
		t.delete(n.children[index], predecessor)
	} else if len(n.children[index+1].keys) >= t.degree {
		successor := t.getSuccessor(n, index)
		n.keys[index] = successor
		t.delete(n.children[index+1], successor)
	} else {
		t.mergeChildren(n, index)
		t.delete(n.children[index], key)
	}
}

// mergeChildren merges two child nodes during deletion.
func (t *BTree[K, V]) mergeChildren(n *node[K, V], index int) {
	leftChild := n.children[index]
	rightChild := n.children[index+1]

	leftChild.keys = append(leftChild.keys, n.keys[index])
	leftChild.values = append(leftChild.values, n.values[index])
	leftChild.keys = append(leftChild.keys, rightChild.keys...)
	leftChild.values = append(leftChild.values, rightChild.values...)
	leftChild.children = append(leftChild.children, rightChild.children...)

	copy(n.keys[index:], n.keys[index+1:])
	copy(n.values[index:], n.values[index+1:])
	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]
	copy(n.children[index+1:], n.children[index+2:])
	n.children = n.children[:len(n.children)-1]
}

// getPredecessor finds the predecessor of a key in the BTree.
func (t *BTree[K, V]) getPredecessor(n *node[K, V], index int) K {
	curr := n.children[index]
	for !curr.leaf {
		curr = curr.children[len(curr.children)-1]
	}
	return curr.keys[len(curr.keys)-1]
}

// getSuccessor finds the successor of a key in the BTree.
func (t *BTree[K, V]) getSuccessor(n *node[K, V], index int) K {
	curr := n.children[index+1]
	for !curr.leaf {
		curr = curr.children[0]
	}
	return curr.keys[0]
}

// search finds the node and index for a given key in the BTree.
func (t *BTree[K, V]) search(n *node[K, V], key K) (*node[K, V], int) {
	index := 0
	for index < len(n.keys) && t.comparator(key, n.keys[index]) > 0 {
		index++
	}
	return n, index
}

// Traversal methods

func (t *BTree[K, V]) inOrderTraversal(n *node[K, V], result *[]collections.Pair[K, V]) {
	if n == nil {
		return
	}

	for i := 0; i < len(n.keys); i++ {
		t.inOrderTraversal(n.children[i], result)
		*result = append(*result, collections.Pair[K, V]{
			Key:   n.keys[i],
			Value: n.values[i],
		})
	}
	t.inOrderTraversal(n.children[len(n.keys)], result)
}

func (t *BTree[K, V]) preOrderTraversal(n *node[K, V], result *[]collections.Pair[K, V]) {
	if n == nil {
		return
	}

	for i := 0; i < len(n.keys); i++ {
		*result = append(*result, collections.Pair[K, V]{
			Key:   n.keys[i],
			Value: n.values[i],
		})
		t.preOrderTraversal(n.children[i], result)
	}
	t.preOrderTraversal(n.children[len(n.keys)], result)
}

func (t *BTree[K, V]) postOrderTraversal(n *node[K, V], result *[]collections.Pair[K, V]) {
	if n == nil {
		return
	}

	for i := 0; i < len(n.keys); i++ {
		t.postOrderTraversal(n.children[i], result)
	}

	for i := 0; i < len(n.keys); i++ {
		*result = append(*result, collections.Pair[K, V]{
			Key:   n.keys[i],
			Value: n.values[i],
		})
	}
	t.postOrderTraversal(n.children[len(n.keys)], result)
}

func (t *BTree[K, V]) levelOrderTraversal(n *node[K, V], result *[]collections.Pair[K, V]) {
	if n == nil {
		return
	}

	queue := []*node[K, V]{n}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for i := 0; i < len(curr.keys); i++ {
			*result = append(*result, collections.Pair[K, V]{
				Key:   curr.keys[i],
				Value: curr.values[i],
			})
			if !curr.leaf {
				queue = append(queue, curr.children[i])
			}
		}
		if !curr.leaf {
			queue = append(queue, curr.children[len(curr.keys)])
		}
	}
}

// Implement Map interface methods

// Put inserts a key-value pair into the BTree.
// If the key already exists, the old value is replaced and returned.
// The boolean return value indicates whether an existing entry was updated.
func (t *BTree[K, V]) Put(key K, value V) (V, bool) {
	err := t.Insert(key, value)
	if err != nil {
		var zero V
		return zero, false
	}
	return value, true
}

// Get retrieves a value from the BTree by its key.
// It returns the value and a boolean indicating whether the key was found.
func (t *BTree[K, V]) Get(key K) (V, bool) {
	node, found := t.Search(key)
	if !found {
		var zero V
		return zero, false
	}
	return node.Value, true
}

// Remove removes a key and its associated value from the BTree.
// It returns the removed value and a boolean indicating whether the key was found.
func (t *BTree[K, V]) Remove(key K) (V, bool) {
	node, found := t.Search(key)
	if !found {
		var zero V
		return zero, false
	}
	value := node.Value
	err := t.Delete(key)
	if err != nil {
		return value, false
	}
	return value, true
}

// ContainsKey checks if the BTree contains the given key.
func (t *BTree[K, V]) ContainsKey(key K) bool {
	_, found := t.Search(key)
	return found
}

// Keys returns a slice of all keys in the BTree.
func (t *BTree[K, V]) Keys() []K {
	keys := make([]K, 0, t.size)
	t.Traverse(tree.InOrder)
	return keys
}

// Values returns a slice of all values in the BTree.
func (t *BTree[K, V]) Values() []V {
	values := make([]V, 0, t.size)
	t.Traverse(tree.InOrder)
	return values
}

// Ensure BTree implements the Map interface
var _ collections.Map[int, int] = (*BTree[int, int])(nil)
