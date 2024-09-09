package tree

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/bits"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/hash"
)

// MerkleTree represents a Merkle Tree data structure implementing the Set interface.
type MerkleTree struct {
	*baseTree[[]byte]
	leaves []*Node[[]byte]
}

// NewMerkleTree creates a new Merkle Tree from the given data.
func NewMerkleTree(data [][]byte) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot create tree with no data")
	}

	hasher, err := hash.NewSipHasher[[]byte]()
	if err != nil {
		return nil, fmt.Errorf("failed to create hasher: %w", err)
	}

	mt := &MerkleTree{
		baseTree: newBaseTree(comp.ByteSliceComparator, hasher),
	}

	if err := mt.Build(data); err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	return mt, nil
}

// Build constructs the Merkle Tree from the given data.
func (mt *MerkleTree) Build(data [][]byte) error {
	if len(data) == 0 {
		return errors.New("cannot build tree with no data")
	}

	mt.leaves = make([]*Node[[]byte], len(data))

	// Create leaf nodes
	for i, item := range data {
		hash := mt.hashData(item)
		mt.leaves[i] = &Node[[]byte]{Value: hash}
	}

	mt.root = mt.buildTree(mt.leaves)
	mt.size = len(mt.leaves)
	return nil
}

// buildTree recursively builds the Merkle Tree from the given nodes.
func (mt *MerkleTree) buildTree(nodes []*Node[[]byte]) *Node[[]byte] {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var nextLevel []*Node[[]byte]

	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		var right *Node[[]byte]
		if i+1 < len(nodes) {
			right = nodes[i+1]
		} else {
			right = &Node[[]byte]{Value: left.Value} // Duplicate last node if odd
		}

		parentHash := mt.hashChildren(left.Value, right.Value)
		parent := &Node[[]byte]{Value: parentHash, Children: []*Node[[]byte]{left, right}}
		nextLevel = append(nextLevel, parent)
	}

	return mt.buildTree(nextLevel)
}

// GetRoot returns the root hash of the Merkle Tree.
func (mt *MerkleTree) GetRoot() []byte {
	if mt.root == nil {
		return nil
	}
	return mt.root.Value
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
			proof = append(proof, sibling.Value)
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

	return comp.ByteSliceComparator(computedHash, rootHash) == 0
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

// getSibling returns the sibling node of the given node.
func (mt *MerkleTree) getSibling(node *Node[[]byte], isRightChild bool) *Node[[]byte] {
	parent := mt.getParent(node)
	if parent == nil {
		return nil
	}
	if isRightChild {
		return parent.Children[0]
	}
	return parent.Children[1]
}

// getParent returns the parent node of the given node.
func (mt *MerkleTree) getParent(node *Node[[]byte]) *Node[[]byte] {
	var findParent func(*Node[[]byte]) *Node[[]byte]
	findParent = func(current *Node[[]byte]) *Node[[]byte] {
		if current == nil || len(current.Children) == 0 {
			return nil
		}
		if current.Children[0] == node || current.Children[1] == node {
			return current
		}
		left := findParent(current.Children[0])
		if left != nil {
			return left
		}
		return findParent(current.Children[1])
	}
	return findParent(mt.root)
}

// Add implements the Set interface.
func (mt *MerkleTree) Add(item []byte) bool {
	if mt.Contains(item) {
		return false
	}
	hash := mt.hashData(item)
	newLeaf := &Node[[]byte]{Value: hash}
	mt.leaves = append(mt.leaves, newLeaf)
	mt.root = mt.buildTree(mt.leaves)
	mt.size++
	return true
}

// Remove implements the Set interface.
func (mt *MerkleTree) Remove(item []byte) bool {
	// Not supported for Merkle Trees
	return false
}

// Contains implements the Set interface.
func (mt *MerkleTree) Contains(item []byte) bool {
	hash := mt.hashData(item)
	for _, leaf := range mt.leaves {
		if comp.ByteSliceComparator(leaf.Value, hash) == 0 {
			return true
		}
	}
	return false
}

// ReverseIterator implements the Iterable interface.
func (mt *MerkleTree) ReverseIterator() collections.Iterator[[]byte] {
	return &merkleReverseIterator{
		currentIndex: len(mt.leaves) - 1,
		tree:         mt,
	}
}

type merkleReverseIterator struct {
	currentIndex int
	tree         *MerkleTree
}

func (it *merkleReverseIterator) HasNext() bool {
	return it.currentIndex >= 0
}

func (it *merkleReverseIterator) Next() []byte {
	if !it.HasNext() {
		panic("no more elements")
	}
	value := it.tree.leaves[it.currentIndex].Value
	it.currentIndex--
	return value
}

// Ensure MerkleTree implements the Set interface
var _ collections.Set[[]byte] = (*MerkleTree)(nil)
