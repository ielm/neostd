package tree

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"math/bits"
	"sync"

	"github.com/ielm/neostd/pkg/collections"
	"github.com/ielm/neostd/pkg/collections/comp"
	"github.com/ielm/neostd/pkg/hash"
)

// MerkleTree represents a Merkle Tree data structure.
type MerkleTree struct {
	root     *MerkleNode
	leaves   []*MerkleNode
	hasher   hash.Hasher[[]byte]
	nodePool *sync.Pool
}

// MerkleNode represents a node in the Merkle Tree.
type MerkleNode struct {
	hash  []byte
	left  *MerkleNode
	right *MerkleNode
}

// NewMerkleTree creates a new Merkle Tree from the given data.
func NewMerkleTree(data [][]byte) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot create tree with no data")
	}

	hasher, err := hash.NewSipHasher[[]byte]()
	if err != nil {
		return nil, err
	}

	tree := &MerkleTree{
		hasher: hasher,
		nodePool: &sync.Pool{
			New: func() interface{} {
				return &MerkleNode{}
			},
		},
	}

	if err := tree.Build(data); err != nil {
		return nil, err
	}

	return tree, nil
}

// Build constructs the Merkle Tree from the given data.
func (mt *MerkleTree) Build(data [][]byte) error {
	if len(data) == 0 {
		return errors.New("cannot build tree with no data")
	}

	mt.leaves = make([]*MerkleNode, len(data))

	// Create leaf nodes
	for i, item := range data {
		hash := mt.hashData(item)
		mt.leaves[i] = mt.newNode(hash, nil, nil)
	}

	mt.root = mt.buildTree(mt.leaves)
	return nil
}

// buildTree recursively builds the Merkle Tree from the given nodes.
func (mt *MerkleTree) buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var nextLevel []*MerkleNode

	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		var right *MerkleNode
		if i+1 < len(nodes) {
			right = nodes[i+1]
		} else {
			right = mt.newNode(left.hash, nil, nil) // Duplicate last node if odd
		}

		parentHash := mt.hashChildren(left.hash, right.hash)
		parent := mt.newNode(parentHash, left, right)
		nextLevel = append(nextLevel, parent)
	}

	return mt.buildTree(nextLevel)
}

// GetRoot returns the root hash of the Merkle Tree.
func (mt *MerkleTree) GetRoot() []byte {
	if mt.root == nil {
		return nil
	}
	return mt.root.hash
}

// GetProof generates a Merkle proof for the data at the given index.
func (mt *MerkleTree) GetProof(index int) ([][]byte, error) {
	if index < 0 || index >= len(mt.leaves) {
		return nil, errors.New("index out of range")
	}

	proof := make([][]byte, 0, bits.Len(uint(len(mt.leaves)-1)))
	current := mt.leaves[index]
	currentIndex := index

	for current != mt.root {
		isRightChild := currentIndex%2 == 1
		sibling := mt.getSibling(current, isRightChild)

		if sibling != nil {
			proof = append(proof, sibling.hash)
		}

		current = mt.getParent(current)
		currentIndex /= 2
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof for the given data and root hash.
func (mt *MerkleTree) VerifyProof(data []byte, proof [][]byte, rootHash []byte) bool {
	computedHash := mt.hashData(data)

	for _, proofElement := range proof {
		computedHash = mt.hashChildren(computedHash, proofElement)
	}

	return bytes.Equal(computedHash, rootHash)
}

// hashData hashes the input data using the tree's hasher.
func (mt *MerkleTree) hashData(data []byte) []byte {
	hash, err := mt.hasher.Hash(data)
	if err != nil {
		// In a production environment, we might want to handle this error more gracefully
		panic(err)
	}
	return hash
}

// hashChildren hashes two child hashes to create a parent hash.
func (mt *MerkleTree) hashChildren(left, right []byte) []byte {
	h := sha256.New()
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}

// newNode creates a new MerkleNode, using the node pool for efficiency.
func (mt *MerkleTree) newNode(hash []byte, left, right *MerkleNode) *MerkleNode {
	node := mt.nodePool.Get().(*MerkleNode)
	node.hash = hash
	node.left = left
	node.right = right
	return node
}

// getSibling returns the sibling node of the given node.
func (mt *MerkleTree) getSibling(node *MerkleNode, isRightChild bool) *MerkleNode {
	parent := mt.getParent(node)
	if parent == nil {
		return nil
	}
	if isRightChild {
		return parent.left
	}
	return parent.right
}

// getParent returns the parent node of the given node.
func (mt *MerkleTree) getParent(node *MerkleNode) *MerkleNode {
	// This is a simplified implementation. In a full implementation,
	// we would maintain parent pointers or use a more sophisticated
	// method to find the parent.
	var findParent func(*MerkleNode) *MerkleNode
	findParent = func(current *MerkleNode) *MerkleNode {
		if current == nil || (current.left != node && current.right != node) {
			return nil
		}
		if current.left == node || current.right == node {
			return current
		}
		left := findParent(current.left)
		if left != nil {
			return left
		}
		return findParent(current.right)
	}
	return findParent(mt.root)
}

// Clear removes all nodes from the tree and returns them to the node pool.
func (mt *MerkleTree) Clear() {
	var clearNode func(*MerkleNode)
	clearNode = func(node *MerkleNode) {
		if node == nil {
			return
		}
		clearNode(node.left)
		clearNode(node.right)
		node.hash = nil
		node.left = nil
		node.right = nil
		mt.nodePool.Put(node)
	}
	clearNode(mt.root)
	mt.root = nil
	mt.leaves = nil
}

// Size returns the number of leaves in the Merkle Tree.
func (mt *MerkleTree) Size() int {
	return len(mt.leaves)
}

// IsEmpty returns true if the Merkle Tree has no nodes.
func (mt *MerkleTree) IsEmpty() bool {
	return mt.root == nil
}

// Iterator returns an iterator for the leaf nodes of the Merkle Tree.
func (mt *MerkleTree) Iterator() collections.Iterator[[]byte] {
	return &merkleTreeIterator{
		tree:  mt,
		index: 0,
	}
}

func (mt *MerkleTree) ReverseIterator() collections.Iterator[[]byte] {
	return &merkleTreeReverseIterator{
		tree:  mt,
		index: len(mt.leaves) - 1,
	}
}

type merkleTreeReverseIterator struct {
	tree  *MerkleTree
	index int
}

func (it *merkleTreeReverseIterator) HasNext() bool {
	return it.index >= 0
}

func (it *merkleTreeReverseIterator) Next() []byte {
	if !it.HasNext() {
		panic("no more elements")
	}
	hash := it.tree.leaves[it.index].hash
	it.index--
	return hash
}

type merkleTreeIterator struct {
	tree  *MerkleTree
	index int
}

func (it *merkleTreeIterator) HasNext() bool {
	return it.index < len(it.tree.leaves)
}

func (it *merkleTreeIterator) Next() []byte {
	if !it.HasNext() {
		panic("no more elements")
	}
	hash := it.tree.leaves[it.index].hash
	it.index++
	return hash
}

// Ensure MerkleTree implements the Collection interface
var _ collections.Collection[[]byte] = (*MerkleTree)(nil)

// Add implements the Collection interface.
func (mt *MerkleTree) Add(item []byte) bool {
	hash := mt.hashData(item)
	newLeaf := mt.newNode(hash, nil, nil)
	mt.leaves = append(mt.leaves, newLeaf)
	mt.root = mt.buildTree(mt.leaves)
	return true
}

// Remove is not efficiently supported for Merkle Trees and is left unimplemented.
func (mt *MerkleTree) Remove(item []byte) bool {
	// Not efficiently supported for Merkle Trees
	return false
}

// Contains checks if the given item's hash is present in the tree's leaves.
func (mt *MerkleTree) Contains(item []byte) bool {
	hash := mt.hashData(item)
	for _, leaf := range mt.leaves {
		if bytes.Equal(leaf.hash, hash) {
			return true
		}
	}
	return false
}

// SetComparator is a no-op for Merkle Trees as they use hash comparisons.
func (mt *MerkleTree) SetComparator(comp comp.Comparator[[]byte]) {
	// No-op for Merkle Trees
}
