package compression

import (
	"strings"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/heap"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/res"
)

// HuffmanNode represents a node in the Huffman tree
type HuffmanNode struct {
	Char  rune
	Freq  int
	Left  *HuffmanNode
	Right *HuffmanNode
}

// HuffmanEncode performs Huffman coding on the input string
func HuffmanEncode(input string) res.Result[map[rune]string] {
	if len(input) == 0 {
		return res.Err[map[rune]string](errors.New(errors.ErrInvalidArgument, "input string is empty"))
	}

	// Count character frequencies
	freqMap := make(map[rune]int)
	for _, char := range input {
		freqMap[char]++
	}

	// Create a min-heap of HuffmanNodes
	h := heap.NewMinBinaryHeap(func(a, b *HuffmanNode) int {
		return a.Freq - b.Freq
	})

	for char, freq := range freqMap {
		h.Push(&HuffmanNode{Char: char, Freq: freq})
	}

	// Build the Huffman tree
	for h.Len() > 1 {
		leftOpt := h.Pop()
		rightOpt := h.Pop()
		if leftOpt.IsNone() || rightOpt.IsNone() {
			return res.Err[map[rune]string](errors.New(errors.ErrInternal, "unexpected empty heap"))
		}
		left := leftOpt.Unwrap()
		right := rightOpt.Unwrap()
		parent := &HuffmanNode{
			Freq:  left.Freq + right.Freq,
			Left:  left,
			Right: right,
		}
		h.Push(parent)
	}

	// Generate Huffman codes
	rootOpt := h.Pop()
	if rootOpt.IsNone() {
		return res.Err[map[rune]string](errors.New(errors.ErrInternal, "unexpected empty heap"))
	}
	root := rootOpt.Unwrap()
	codeMap := make(map[rune]string)
	generateCodes(root, "", codeMap)

	return res.Ok(codeMap)
}

// generateCodes recursively generates Huffman codes for each character
func generateCodes(node *HuffmanNode, code string, codeMap map[rune]string) {
	if node == nil {
		return
	}
	if node.Left == nil && node.Right == nil {
		codeMap[node.Char] = code
		return
	}
	generateCodes(node.Left, code+"0", codeMap)
	generateCodes(node.Right, code+"1", codeMap)
}

// HuffmanDecode decodes a Huffman-encoded string
func HuffmanDecode(encoded string, codeMap map[rune]string) res.Result[string] {
	if len(encoded) == 0 {
		return res.Err[string](errors.New(errors.ErrInvalidArgument, "encoded string is empty"))
	}
	if len(codeMap) == 0 {
		return res.Err[string](errors.New(errors.ErrInvalidArgument, "codeMap is empty"))
	}

	// Create a reverse lookup map
	reverseMap := make(map[string]rune)
	for char, code := range codeMap {
		reverseMap[code] = char
	}

	var decoded strings.Builder
	currentCode := ""
	for _, bit := range encoded {
		currentCode += string(bit)
		if char, found := reverseMap[currentCode]; found {
			decoded.WriteRune(char)
			currentCode = ""
		}
	}

	if currentCode != "" {
		return res.Err[string](errors.New(errors.ErrInvalidArgument, "invalid encoded string"))
	}

	return res.Ok(decoded.String())
}

// HuffmanCompressor implements the Compressor interface for Huffman coding
type HuffmanCompressor struct{}

func (hc *HuffmanCompressor) Compress(input string) res.Result[collections.Pair[string, map[rune]string]] {
	codeMapResult := HuffmanEncode(input)
	if codeMapResult.IsErr() {
		return res.Err[collections.Pair[string, map[rune]string]](codeMapResult.UnwrapErr())
	}
	codeMap := codeMapResult.Unwrap()

	var compressed strings.Builder
	for _, char := range input {
		compressed.WriteString(codeMap[char])
	}

	return res.Ok(collections.Pair[string, map[rune]string]{
		Key:   compressed.String(),
		Value: codeMap,
	})
}

func (hc *HuffmanCompressor) Decompress(compressed collections.Pair[string, map[rune]string]) res.Result[string] {
	return HuffmanDecode(compressed.Key, compressed.Value)
}

// HuffmanIterator implements the Iterator interface for Huffman coding
type HuffmanIterator struct {
	input    collections.Iterator[string]
	codeMap  map[rune]string
	buffer   string
	position int
}

func NewHuffmanIterator(input collections.Iterator[string]) *HuffmanIterator {
	return &HuffmanIterator{
		input:   input,
		codeMap: make(map[rune]string),
	}
}

func (hi *HuffmanIterator) HasNext() bool {
	return hi.position < len(hi.buffer) || hi.input.HasNext()
}

func (hi *HuffmanIterator) Next() res.Option[string] {
	if hi.position >= len(hi.buffer) {
		if !hi.input.HasNext() {
			return res.None[string]()
		}
		chunkOpt := hi.input.Next()
		if chunkOpt.IsNone() {
			return res.None[string]()
		}
		chunk := chunkOpt.Unwrap()

		codeMapResult := HuffmanEncode(chunk)
		if codeMapResult.IsErr() {
			return res.None[string]()
		}
		hi.codeMap = codeMapResult.Unwrap()

		var compressed strings.Builder
		for _, char := range chunk {
			compressed.WriteString(hi.codeMap[char])
		}
		hi.buffer = compressed.String()
		hi.position = 0
	}

	result := hi.buffer[hi.position]
	hi.position++
	return res.Some(string(result))
}

// HuffmanCompressIterator compresses an iterator of strings using Huffman coding
func CompressIterator(input collections.Iterator[string]) res.Result[collections.Iterator[string]] {
	return res.Ok(collections.Iterator[string](NewHuffmanIterator(input)))
}

// HuffmanDecompressIterator decompresses an iterator of Huffman-encoded strings
func DecompressIterator(input collections.Iterator[string], codeMap map[rune]string) res.Result[collections.Iterator[string]] {
	return res.Ok(collections.Iterator[string](&HuffmanDecompressIterator{
		input:   input,
		codeMap: codeMap,
	}))
}

type HuffmanDecompressIterator struct {
	input   collections.Iterator[string]
	codeMap map[rune]string
	buffer  string
}

func (hdi *HuffmanDecompressIterator) HasNext() bool {
	return len(hdi.buffer) > 0 || hdi.input.HasNext()
}

func (hdi *HuffmanDecompressIterator) Next() res.Option[string] {
	if len(hdi.buffer) == 0 {
		if !hdi.input.HasNext() {
			return res.None[string]()
		}
		chunkOpt := hdi.input.Next()
		if chunkOpt.IsNone() {
			return res.None[string]()
		}
		chunk := chunkOpt.Unwrap()

		decodedResult := HuffmanDecode(chunk, hdi.codeMap)
		if decodedResult.IsErr() {
			return res.None[string]()
		}
		hdi.buffer = decodedResult.Unwrap()
	}

	result := string(hdi.buffer[0])
	hdi.buffer = hdi.buffer[1:]
	return res.Some(result)
}
