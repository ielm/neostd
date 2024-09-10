package rabinkarp

import (
	"strings"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/res"
)

const (
	prime   = 101
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

// RabinKarp performs the Rabin-Karp string searching algorithm
// It returns a Result containing a map of patterns to their occurrences in the text
func RabinKarp(text string, patterns []string) res.Result[map[string][]int] {
	if len(patterns) == 0 {
		return res.Ok(make(map[string][]int))
	}

	// Find the longest pattern length
	maxLen := 0
	for _, pattern := range patterns {
		if len(pattern) > maxLen {
			maxLen = len(pattern)
		}
	}

	if maxLen == 0 {
		return res.Ok(make(map[string][]int))
	}

	if len(text) < maxLen {
		return res.Ok(make(map[string][]int))
	}

	// Precompute hash values for patterns
	patternHashes := make(map[uint][]string)
	for _, pattern := range patterns {
		hash := computeHash(pattern)
		patternHashes[hash] = append(patternHashes[hash], pattern)
	}

	result := make(map[string][]int)
	textHash := computeHash(text[:maxLen])

	// Sliding window approach
	for i := 0; i <= len(text)-maxLen; i++ {
		if patterns, ok := patternHashes[textHash]; ok {
			for _, pattern := range patterns {
				if text[i:i+len(pattern)] == pattern {
					result[pattern] = append(result[pattern], i)
				}
			}
		}

		if i < len(text)-maxLen {
			textHash = updateHash(textHash, text[i], text[i+maxLen], maxLen)
		}
	}

	return res.Ok(result)
}

// computeHash calculates the initial hash value for a string
func computeHash(s string) uint {
	var hash uint
	for i := 0; i < len(s); i++ {
		hash = (hash*uint(prime) + uint(s[i])) % maxUint
	}
	return hash
}

// updateHash updates the hash value for the sliding window
func updateHash(prevHash uint, oldChar byte, newChar byte, patternLen int) uint {
	hash := prevHash
	hash = hash - uint(oldChar)*pow(uint(prime), uint(patternLen-1))
	hash = (hash*uint(prime) + uint(newChar)) % maxUint
	return hash
}

// pow calculates (base^exp) % maxUint efficiently
func pow(base, exp uint) uint {
	result := uint(1)
	for exp > 0 {
		if exp&1 == 1 {
			result = (result * base) % maxUint
		}
		base = (base * base) % maxUint
		exp >>= 1
	}
	return result
}

// RabinKarpWithOptions performs the Rabin-Karp string searching algorithm with additional options
func RabinKarpWithOptions(text string, patterns []string, options ...RabinKarpOption) res.Result[map[string][]int] {
	config := defaultRabinKarpConfig()
	for _, option := range options {
		option(config)
	}

	if !config.caseSensitive {
		text = config.toLower(text)
		for i, pattern := range patterns {
			patterns[i] = config.toLower(pattern)
		}
	}

	return RabinKarp(text, patterns)
}

// RabinKarpConfig holds configuration options for the Rabin-Karp algorithm
type RabinKarpConfig struct {
	caseSensitive bool
	toLower       func(string) string
}

// RabinKarpOption is a function type for setting Rabin-Karp options
type RabinKarpOption func(*RabinKarpConfig)

// defaultRabinKarpConfig returns the default configuration for Rabin-Karp
func defaultRabinKarpConfig() *RabinKarpConfig {
	return &RabinKarpConfig{
		caseSensitive: true,
		toLower:       strings.ToLower,
	}
}

// WithCaseInsensitive sets the Rabin-Karp algorithm to be case-insensitive
func WithCaseInsensitive() RabinKarpOption {
	return func(c *RabinKarpConfig) {
		c.caseSensitive = false
	}
}

// WithCustomLowerCase sets a custom lowercase function for case-insensitive matching
func WithCustomLowerCase(lowerFunc func(string) string) RabinKarpOption {
	return func(c *RabinKarpConfig) {
		c.toLower = lowerFunc
	}
}

// RabinKarpIterator performs the Rabin-Karp algorithm on an iterator of text chunks
func RabinKarpIterator(textIter collections.Iterator[string], patterns []string) res.Result[map[string][]int] {
	result := make(map[string][]int)
	offset := 0

	for textIter.HasNext() {
		chunkOpt := textIter.Next()
		if chunkOpt.IsNone() {
			return res.Err[map[string][]int](errors.New(errors.ErrInvalidArgument, "invalid text iterator"))
		}

		chunk := chunkOpt.Unwrap()
		chunkResult := RabinKarp(chunk, patterns)
		if chunkResult.IsErr() {
			return res.Err[map[string][]int](chunkResult.UnwrapErr())
		}

		chunkMatches := chunkResult.Unwrap()
		for pattern, positions := range chunkMatches {
			for _, pos := range positions {
				result[pattern] = append(result[pattern], pos+offset)
			}
		}

		offset += len(chunk)
	}

	return res.Ok(result)
}
