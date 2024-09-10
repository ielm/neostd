// A Trie, also called a prefix tree, is an efficient tree-like data structure used
// to store and retrieve strings. Each node in the trie represents a character, and
// the path from the root to a node forms a prefix of one or more strings.
//
// Key features of a Trie:
//  1. The root node is typically empty.
//  2. Each node stores a character and has multiple children, one for each possible next character.
//  3. The paths from root to leaf represent complete words.
//  4. Common prefixes are shared, making tries memory-efficient for storing many strings with similar prefixes.
//  5. Operations like insertion, deletion, and search have a time complexity of O(m),
//     where m is the length of the string, regardless of the number of strings in the trie.
//
// ```plaintext
//		   root
//        /    \
//       h      b
//      / \      \
//     e   i*     a
//    / \        / \
//   a   n*     r   i
//  / \        / \   \
// d*  r*     *   e*  n*
//                     \
//                      t*
// ```

package tree

import (
	"unicode/utf8"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/collections/maps"
	"github.com/ielm/neostd/errors"
)

// Trie represents a trie data structure.
// This implementation uses a HashMap for storing children, allowing for efficient
// operations on Unicode strings. It implements the Set interface for strings.
type Trie struct {
	*baseTree[string]
	root *trieNode
}

// trieNode represents a single node in the Trie.
type trieNode struct {
	children *maps.HashMap[rune, *trieNode] // Map of child nodes
	isEnd    bool                           // Flag indicating if this node represents the end of a word
}

// NewTrie creates a new Trie.
//
// This function initializes an empty Trie with a root node.
//
// Example:
//
//	trie := NewTrie()
func NewTrie() *Trie {
	return &Trie{
		baseTree: newBaseTree(comp.GenericComparator[string](), nil),
		root:     newTrieNode(),
	}
}

// newTrieNode creates a new trieNode.
func newTrieNode() *trieNode {
	return &trieNode{
		children: maps.NewHashMap[rune, *trieNode](comp.GenericComparator[rune]()),
	}
}

// Insert adds a word to the trie.
//
// This method inserts a word into the trie, creating new nodes as necessary.
// If the word already exists, the method does nothing.
//
// Example:
//
//	trie.Insert("hello")
func (t *Trie) Insert(word string) error {
	if word == "" {
		return errors.New(errors.ErrInvalidArgument, "cannot insert empty string")
	}

	node := t.root
	for _, ch := range word {
		if child, exists := node.children.Get(ch); exists {
			node = child
		} else {
			newNode := newTrieNode()
			node.children.Put(ch, newNode)
			node = newNode
		}
	}
	if !node.isEnd {
		node.isEnd = true
		t.size++
	}
	return nil
}

// Search checks if a word exists in the trie.
//
// This method traverses the trie to check if the given word is present.
// It returns true if the word is found, false otherwise.
//
// Example:
//
//	found := trie.Search("hello")
func (t *Trie) Search(word string) bool {
	node := t.findNode(word)
	return node != nil && node.isEnd
}

// StartsWith checks if any word in the trie starts with the given prefix.
//
// This method traverses the trie to check if there's any word with the given prefix.
// It returns true if a prefix is found, false otherwise.
//
// Example:
//
//	hasPrefix := trie.StartsWith("he")
func (t *Trie) StartsWith(prefix string) bool {
	return t.findNode(prefix) != nil
}

// findNode is a helper method that finds the node corresponding to a given string.
func (t *Trie) findNode(s string) *trieNode {
	node := t.root
	for _, ch := range s {
		if child, exists := node.children.Get(ch); exists {
			node = child
		} else {
			return nil
		}
	}
	return node
}

// Delete removes a word from the trie.
//
// This method removes the given word from the trie if it exists.
// It uses a depth-first search approach to remove nodes that are no longer needed.
//
// Example:
//
//	err := trie.Delete("hello")
func (t *Trie) Delete(word string) error {
	if word == "" {
		return errors.New(errors.ErrInvalidArgument, "cannot delete empty string")
	}

	var dfs func(node *trieNode, s string, depth int) bool
	dfs = func(node *trieNode, s string, depth int) bool {
		if depth == len(s) {
			if !node.isEnd {
				return false
			}
			node.isEnd = false
			t.size--
			return node.children.IsEmpty()
		}

		ch, _ := utf8.DecodeRuneInString(s[depth:])
		child, exists := node.children.Get(ch)
		if !exists {
			return false
		}

		shouldDeleteChild := dfs(child, s, depth+1)
		if shouldDeleteChild {
			node.children.Remove(ch)
			return node.children.IsEmpty() && !node.isEnd
		}
		return false
	}

	dfs(t.root, word, 0)
	return nil
}

// Clear removes all words from the trie.
//
// This method resets the trie to its initial state.
//
// Example:
//
//	trie.Clear()
func (t *Trie) Clear() {
	t.root = newTrieNode()
	t.size = 0
}

// Words returns all words in the trie.
//
// This method performs a depth-first search to collect all words in the trie.
//
// Example:
//
//	words := trie.Words()
func (t *Trie) Words() []string {
	var result []string
	var dfs func(node *trieNode, current []rune)
	dfs = func(node *trieNode, current []rune) {
		if node.isEnd {
			result = append(result, string(current))
		}
		node.children.ForEach(func(ch rune, child *trieNode) {
			dfs(child, append(current, ch))
		})
	}
	dfs(t.root, []rune{})
	return result
}

// Iterator returns an iterator over the words in the trie.
//
// The iterator allows forward traversal of all words in the trie.
//
// Example:
//
//	it := trie.Iterator()
//	for it.HasNext() {
//		fmt.Println(it.Next())
//	}
func (t *Trie) Iterator() collections.Iterator[string] {
	return &trieIterator{
		trie:  t,
		words: t.Words(),
		index: 0,
	}
}

// ReverseIterator returns a reverse iterator over the words in the trie.
//
// The reverse iterator allows backward traversal of all words in the trie.
//
// Example:
//
//	it := trie.ReverseIterator()
//	for it.HasNext() {
//		fmt.Println(it.Next())
//	}
func (t *Trie) ReverseIterator() collections.Iterator[string] {
	words := t.Words()
	return &trieIterator{
		trie:    t,
		words:   words,
		index:   len(words) - 1,
		reverse: true,
	}
}

// trieIterator is an iterator for the Trie.
type trieIterator struct {
	trie    *Trie
	words   []string
	index   int
	reverse bool
}

// HasNext checks if there are more elements in the iterator.
func (it *trieIterator) HasNext() bool {
	if it.reverse {
		return it.index >= 0
	}
	return it.index < len(it.words)
}

// Next returns the next element in the iterator.
func (it *trieIterator) Next() string {
	if !it.HasNext() {
		panic("no more elements")
	}
	word := it.words[it.index]
	if it.reverse {
		it.index--
	} else {
		it.index++
	}
	return word
}

// Add implements the Set interface.
//
// This method adds a word to the trie. It returns true if the word was added,
// false if it already existed.
//
// Example:
//
//	added := trie.Add("hello")
func (t *Trie) Add(word string) bool {
	if t.Search(word) {
		return false // Word already exists
	}
	t.Insert(word)
	return true
}

// Remove implements the Set interface.
//
// This method removes a word from the trie. It returns true if the word was removed,
// false if it didn't exist.
//
// Example:
//
//	removed := trie.Remove("hello")
func (t *Trie) Remove(word string) bool {
	if !t.Search(word) {
		return false // Word doesn't exist
	}
	t.Delete(word)
	return true
}

// Contains implements the Set interface.
//
// This method checks if a word exists in the trie.
//
// Example:
//
//	exists := trie.Contains("hello")
func (t *Trie) Contains(word string) bool {
	return t.Search(word)
}

// SetComparator is a no-op for Trie as it uses string comparison by default.
func (t *Trie) SetComparator(comp comp.Comparator[string]) {
	// No-op
}

// Ensure Trie implements the Set interface
var _ collections.Set[string] = (*Trie)(nil)
