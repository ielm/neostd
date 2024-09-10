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
	"github.com/ielm/neostd/res"
)

// Trie represents a generic trie data structure.
type Trie[T any] struct {
	*baseTree[string]
	root *trieNode[T]
}

// trieNode represents a single node in the Trie.
type trieNode[T any] struct {
	children *maps.HashMap[rune, *trieNode[T]]
	value    *T
	isEnd    bool
}

// NewTrie creates a new Trie.
func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{
		baseTree: newBaseTree(comp.GenericComparator[string](), nil),
		root:     newTrieNode[T](),
	}
}

// newTrieNode creates a new trieNode.
func newTrieNode[T any]() *trieNode[T] {
	return &trieNode[T]{
		children: maps.NewHashMap[rune, *trieNode[T]](comp.GenericComparator[rune]()),
	}
}

// Insert adds a word to the trie with an associated value.
func (t *Trie[T]) Insert(word string, value T) error {
	if word == "" {
		return errors.New(errors.ErrInvalidArgument, "cannot insert empty string")
	}

	node := t.root
	for _, ch := range word {
		if child, exists := node.children.Get(ch); exists {
			node = child
		} else {
			newNode := newTrieNode[T]()
			node.children.Put(ch, newNode)
			node = newNode
		}
	}
	if !node.isEnd {
		node.isEnd = true
		node.value = &value
		t.size++
	}
	return nil
}

// Search checks if a word exists in the trie and returns its value.
func (t *Trie[T]) Search(word string) (*T, bool) {
	node := t.findNode(word)
	if node != nil && node.isEnd {
		return node.value, true
	}
	return nil, false
}

// StartsWith checks if any word in the trie starts with the given prefix.
func (t *Trie[T]) StartsWith(prefix string) bool {
	return t.findNode(prefix) != nil
}

// findNode is a helper method that finds the node corresponding to a given string.
func (t *Trie[T]) findNode(s string) *trieNode[T] {
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
func (t *Trie[T]) Delete(word string) error {
	if word == "" {
		return errors.New(errors.ErrInvalidArgument, "cannot delete empty string")
	}

	var dfs func(node *trieNode[T], s string, depth int) bool
	dfs = func(node *trieNode[T], s string, depth int) bool {
		if depth == len(s) {
			if !node.isEnd {
				return false
			}
			node.isEnd = false
			node.value = nil
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
func (t *Trie[T]) Clear() {
	t.root = newTrieNode[T]()
	t.size = 0
}

// Words returns all words in the trie.
func (t *Trie[T]) Words() []string {
	var result []string
	var dfs func(node *trieNode[T], current []rune)
	dfs = func(node *trieNode[T], current []rune) {
		if node.isEnd {
			result = append(result, string(current))
		}
		node.children.ForEach(func(ch rune, child *trieNode[T]) {
			dfs(child, append(current, ch))
		})
	}
	dfs(t.root, []rune{})
	return result
}

// Iterator returns an iterator over the words in the trie.
func (t *Trie[T]) Iterator() collections.Iterator[string] {
	return &trieIterator[T]{
		trie:  t,
		words: t.Words(),
		index: 0,
	}
}

// ReverseIterator returns a reverse iterator over the words in the trie.
func (t *Trie[T]) ReverseIterator() collections.Iterator[string] {
	words := t.Words()
	return &trieIterator[T]{
		trie:    t,
		words:   words,
		index:   len(words) - 1,
		reverse: true,
	}
}

// trieIterator is an iterator for the Trie.
type trieIterator[T any] struct {
	trie    *Trie[T]
	words   []string
	index   int
	reverse bool
}

// HasNext checks if there are more elements in the iterator.
func (it *trieIterator[T]) HasNext() bool {
	if it.reverse {
		return it.index >= 0
	}
	return it.index < len(it.words)
}

// Next returns the next element in the iterator.
func (it *trieIterator[T]) Next() res.Option[string] {
	if !it.HasNext() {
		return res.None[string]()
	}
	word := it.words[it.index]
	if it.reverse {
		it.index--
	} else {
		it.index++
	}
	return res.Some(word)
}

// Add implements the Set interface.
func (t *Trie[T]) Add(word string) bool {
	if _, found := t.Search(word); found {
		return false // Word already exists
	}
	var zero T
	t.Insert(word, zero)
	return true
}

// Remove implements the Set interface.
func (t *Trie[T]) Remove(word string) bool {
	if _, found := t.Search(word); !found {
		return false // Word doesn't exist
	}
	t.Delete(word)
	return true
}

// Contains implements the Set interface.
func (t *Trie[T]) Contains(word string) bool {
	_, found := t.Search(word)
	return found
}

// SetComparator is a no-op for Trie as it uses string comparison by default.
func (t *Trie[T]) SetComparator(comp comp.Comparator[string]) {
	// No-op
}

// Ensure Trie implements the Set interface
var _ collections.Set[string] = (*Trie[any])(nil)
