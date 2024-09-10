package kmp

import (
	"strings"

	"github.com/ielm/neostd/res"
)

// KMP performs the Knuth-Morris-Pratt string matching algorithm
// It returns a Result containing an array of indices where the pattern is found in the text
// time complexity is O(M + N) where M is the length of the pattern and N is the length of the text
func KMP(text, pattern string) res.Result[[]int] {
	if len(pattern) == 0 {
		return res.Ok([]int{})
	}

	if len(text) == 0 {
		return res.Ok([]int{})
	}

	// Compute the failure function (equivalent to LPS array in traditional KMP)
	failure := computeFailureFunction(pattern)

	var matches []int
	j := 0 // Index for pattern
	// Main KMP algorithm loop
	for i := 0; i < len(text); {
		if text[i] == pattern[j] {
			// Characters match, move both pointers
			if j == len(pattern)-1 {
				// Full pattern match found
				matches = append(matches, i-j)
				// Use failure function to shift pattern
				j = failure[j]
			} else {
				j++
			}
			i++
		} else if j > 0 {
			// Mismatch after some matching characters
			// Use failure function to shift pattern
			j = failure[j-1]
		} else {
			// No match found, move to next character in text
			i++
		}
	}

	return res.Ok(matches)
}

// computeFailureFunction computes the failure function for the KMP algorithm
// This is equivalent to the LPS (Longest Proper Prefix which is also Suffix) array in traditional KMP
func computeFailureFunction(pattern string) []int {
	failure := make([]int, len(pattern))
	failure[0] = 0
	j := 0
	// Compute failure function for each character in the pattern
	for i := 1; i < len(pattern); {
		if pattern[i] == pattern[j] {
			// Characters match, increment length of matched substring
			j++
			failure[i] = j
			i++
		} else if j > 0 {
			// Mismatch after some matching characters
			// Fall back using the failure function
			j = failure[j-1]
		} else {
			// No match found
			failure[i] = 0
			i++
		}
	}
	return failure
}

// KMPWithOptions performs the Knuth-Morris-Pratt string matching algorithm with additional options
func KMPWithOptions(text, pattern string, options ...KMPOption) res.Result[[]int] {
	config := defaultKMPConfig()
	for _, option := range options {
		option(config)
	}

	if config.caseSensitive {
		return KMP(text, pattern)
	}

	return KMP(config.toLower(text), config.toLower(pattern))
}

// KMPConfig holds configuration options for the KMP algorithm
type KMPConfig struct {
	caseSensitive bool
	toLower       func(string) string
}

// KMPOption is a function type for setting KMP options
type KMPOption func(*KMPConfig)

// defaultKMPConfig returns the default configuration for KMP
func defaultKMPConfig() *KMPConfig {
	return &KMPConfig{
		caseSensitive: true,
		toLower:       strings.ToLower,
	}
}

// WithCaseInsensitive sets the KMP algorithm to be case-insensitive
func WithCaseInsensitive() KMPOption {
	return func(c *KMPConfig) {
		c.caseSensitive = false
	}
}

// WithCustomLowerCase sets a custom lowercase function for case-insensitive matching
func WithCustomLowerCase(lowerFunc func(string) string) KMPOption {
	return func(c *KMPConfig) {
		c.toLower = lowerFunc
	}
}
