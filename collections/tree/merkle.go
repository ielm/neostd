package tree

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/bits"
	"sync"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/hash"
)

// MerkleTree represents a Merkle Tree data structure implementing the Set interface.
type MerkleTree struct {
	*baseTree[[]byte]
	leaves     []*Node[[]byte]
	levelCount int
	hasher     *hash.SipHasher
	mu         sync.RWMutex
}

// NewMerkleTree creates a new Merkle Tree from the given data.
func NewMerkleTree(data [][]byte) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New(errors.ErrInvalidArgument, "cannot create tree with no data")
	}

	hasher, err := hash.NewSipHasher()
	if err != nil {
		return nil, fmt.Errorf("failed to create SipHasher: %w", err)
	}

	mt := &MerkleTree{
		baseTree: newBaseTree(comp.ByteSliceComparator, hasher),
		hasher:   hasher,
	}

	if err := mt.Build(data); err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	return mt, nil
}

// NewWithHasher creates a new Merkle Tree with a custom SipHasher.
func NewWithHasher(data [][]byte, hasher *hash.SipHasher) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New(errors.ErrInvalidArgument, "cannot create tree with no data")
	}

	mt := &MerkleTree{
		baseTree: newBaseTree(comp.ByteSliceComparator, hasher),
		hasher:   hasher,
	}

	if err := mt.Build(data); err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	return mt, nil
}

// Build constructs the Merkle Tree from the given data.
func (mt *MerkleTree) Build(data [][]byte) error {
	if len(data) == 0 {
		return errors.New(errors.ErrInvalidArgument, "cannot build tree with no data")
	}

	mt.leaves = make([]*Node[[]byte], len(data))

	// Create leaf nodes
	for i, item := range data {
		hash := mt.hashData(item)
		mt.leaves[i] = &Node[[]byte]{Value: hash}
	}

	mt.root = mt.buildTree(mt.leaves)
	mt.size = len(mt.leaves)
	mt.levelCount = mt.calculateLevelCount(len(mt.leaves))
	return nil
}

// calculateLevelCount calculates the number of levels in the tree
func (mt *MerkleTree) calculateLevelCount(leafCount int) int {
	return bits.Len(uint(leafCount - 1))
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
		return nil, errors.New(errors.ErrOutOfBounds, "index out of range")
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

// Update updates the value at the given index and recalculates the affected hashes.
func (mt *MerkleTree) Update(index int, newData []byte) error {
	if index < 0 || index >= len(mt.leaves) {
		return errors.New(errors.ErrOutOfBounds, "index out of range")
	}

	newHash := mt.hashData(newData)
	mt.leaves[index].Value = newHash

	current := mt.leaves[index]
	currentIndex := index
	level := 0

	for level < mt.levelCount {
		isRightChild := currentIndex%2 == 1
		sibling := mt.getSibling(current, isRightChild)
		parent := mt.getParent(current)

		if isRightChild {
			parent.Value = mt.hashChildren(sibling.Value, current.Value)
		} else {
			parent.Value = mt.hashChildren(current.Value, sibling.Value)
		}

		current = parent
		currentIndex /= 2
		level++
	}

	return nil
}

// Diff returns the indices of leaves that differ between this tree and another.
func (mt *MerkleTree) Diff(other *MerkleTree) ([]int, error) {
	if len(mt.leaves) != len(other.leaves) {
		return nil, errors.New(errors.ErrInvalidArgument, "trees have different sizes")
	}

	diffIndices := []int{}
	queue := [][2]*Node[[]byte]{{mt.root, other.root}}
	index := 0
	levelSize := 1

	for len(queue) > 0 {
		node1, node2 := queue[0][0], queue[0][1]
		queue = queue[1:]

		if !bytes.Equal(node1.Value, node2.Value) {
			if len(node1.Children) == 0 {
				diffIndices = append(diffIndices, index)
			} else {
				queue = append(queue, [2]*Node[[]byte]{node1.Children[0], node2.Children[0]})
				queue = append(queue, [2]*Node[[]byte]{node1.Children[1], node2.Children[1]})
			}
		}

		index++
		if index == levelSize {
			index = 0
			levelSize *= 2
		}
	}

	return diffIndices, nil
}

// hashData now uses the SipHasher
func (mt *MerkleTree) hashData(data []byte) []byte {
	mt.hasher.Reset()
	mt.hasher.Write(data)
	return mt.hasher.Sum(nil)
}

// hashChildren now uses the SipHasher
func (mt *MerkleTree) hashChildren(left, right []byte) []byte {
	mt.hasher.Reset()
	mt.hasher.Write(left)
	mt.hasher.Write(right)
	return mt.hasher.Sum(nil)
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

// Add implements efficient insertion
func (mt *MerkleTree) Add(item []byte) bool {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	if mt.Contains(item) {
		return false
	}
	hash := mt.hashData(item)
	newLeaf := &Node[[]byte]{Value: hash}
	mt.leaves = append(mt.leaves, newLeaf)
	mt.size++
	mt.rebalance()
	return true
}

// Remove implements element deletion
func (mt *MerkleTree) Remove(item []byte) bool {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	hash := mt.hashData(item)
	for i, leaf := range mt.leaves {
		if comp.ByteSliceComparator(leaf.Value, hash) == 0 {
			mt.leaves = append(mt.leaves[:i], mt.leaves[i+1:]...)
			mt.size--
			mt.rebalance()
			return true
		}
	}
	return false
}

// rebalance rebuilds the tree after insertion or deletion
func (mt *MerkleTree) rebalance() {
	mt.root = mt.buildTree(mt.leaves)
	mt.levelCount = mt.calculateLevelCount(len(mt.leaves))
}

// Contains implements the Set interface.
func (mt *MerkleTree) Contains(item []byte) bool {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
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

// Serialize the MerkleTree
func (mt *MerkleTree) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(mt)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize a MerkleTree
func DeserializeMerkleTree(data []byte) (*MerkleTree, error) {
	var mt MerkleTree
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&mt)
	if err != nil {
		return nil, err
	}
	return &mt, nil
}
