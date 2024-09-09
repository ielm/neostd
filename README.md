# neostd

neostd is a flexible toolkit providing a collection of useful data structures, algorithms, and utilities implemented in Go. This project aims to create a comprehensive set of tools that can be easily integrated into various Go applications.

## Table of Contents

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Installation

To use neostd in your Go project, you can install it using `go get`:

```bash
go get github.com/ielm/neostd
```

## Features

neostd currently includes the following components

### Collections

- **Vector**: A dynamic array implementation
- **LinkedList**: A doubly linked list
- **HashMap**: A hash table implementation
- **BinaryHeap**: A priority queue implemented as a binary heap
- **VecDeque**: A double-ended queue implemented with a growable ring buffer

### Caching

- **Cache**: A generic caching system with support for various eviction policies:
  - LRU (Least Recently Used)
  - LFU (Least Frequently Used)
  - LFRU (Least Frequently/Recently Used)

### Probabilistic Data Structures

- **BloomFilter**: Space-efficient probabilistic data structure for set membership testing
- **CuckooFilter**: Space-efficient probabilistic data structure with support for deletions
- **XorFilter**: Another space-efficient probabilistic data structure for set membership testing

### Hashing

- **SipHasher**: Implementation of the SipHash algorithm

### Utilities

- **Comparators**: Generic comparison functions for ordered types

## Usage

Here are some examples of how to use neostd components:

### Vector

```go
import "github.com/ielm/neostd/pkg/collections/vector"

// Create a new vector
vec := vector.VecWithCapacity(10, collections.IntComparator)

// Add elements
vec.Push(1)
vec.Push(2)
vec.Push(3)

// Access elements
firstElement, := vec.Get(0)
fmt.Println(firstElement) // Output: 1

// Remove last element
lastElement, := vec.Pop()
fmt.Println(lastElement) // Output: 3

```

### HashMap

```go
import "github.com/ielm/neostd/pkg/collections/maps"

// Create a new HashMap
hm := maps.NewHashMap(string, int)

// Add key-value pairs
hm.Put("one", 1)
hm.Put("two", 2)

// Get a value
value, exists := hm.Get("one")
if exists {
    fmt.Println(value) // Output: 1
}

// Remove a key-value pair
removedValue, removed := hm.Remove("two")
if removed {
    fmt.Println(removedValue) // Output: 2
}

```

### BloomFilter

```go
import "github.com/ielm/neostd/pkg/collections/filter"

// Create a new Bloom filter
bf, := filter.NewBloomFilter(1000, 0.01)

// Add elements
bf.Add([]byte("example"))

// Check for membership
if bf.Contains([]byte("example")) {
    fmt.Println("Element might be in the set")
}
```

### CuckooFilter

```go
import "github.com/ielm/neostd/pkg/collections/filter"

// Create a new Cuckoo filter
cf, := filter.NewCuckooFilter(1000, 0.01)

// Add elements
cf.Add([]byte("example"))

// Check for membership
if cf.Contains([]byte("example")) {
    fmt.Println("Element might be in the set")
}
```

### XorFilter

```go
import "github.com/ielm/neostd/pkg/collections/filter"

// Create a new Xor filter
xf, := filter.NewXorFilter(1000, 0.01)

// Add elements
xf.Add([]byte("example"))

// Check for membership
if xf.Contains([]byte("example")) {
    fmt.Println("Element might be in the set")
}
```

### Cache

```go
import (
    "github.com/ielm/neostd/pkg/collections/cache"
    "github.com/ielm/neostd/pkg/collections"
)
// Create a new Cache with LRU policy
lruCache := cache.NewCache(1000, cache.NewLRUOrderPolicy(1000), collections.StringComparator)

// Add key-value pairs
lruCache.Set("key1", "value1")
lruCache.Set("key2", "value2")

// Get a value
value, exists := lruCache.Get("key1")
if exists {
    fmt.Println(value) // Output: value1
}

// Remove a key-value pair
lruCache.Remove("key2")

// Create a new Cache with LFU policy
lfuCache := cache.NewCache(1000, cache.NewLFUOrderPolicy(1000), collections.StringComparator)

// Create a new Cache with LFRU policy
lfruCache := cache.NewCache(1000, cache.NewLFRUOrderPolicy(1000), collections.StringComparator)

// Clear the cache
lruCache.Clear()

// Get the size of the cache
size := lruCache.Size()
fmt.Println(size) // Output: 0
```

### SipHasher

```go
import "github.com/ielm/neostd/pkg/hash"

// Create a new SipHasher
hasher, err := hash.NewSipHasher(128)
if err != nil {
    panic(err)
}

// Hash a string
hash := hasher.Hash([]byte("example"))
fmt.Println(hash) // Output: [128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128]
```

For more detailed usage examples, please refer to the documentation of each package.

## Contributing

Contributions to neostd are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Note: This project is a work in progress. New features and improvements will be added over time. Please check back regularly for updates.
