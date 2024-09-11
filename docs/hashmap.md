# HashMap

A hash map implementation using quadratic probing and SIMD-like lookup, providing O(1) get and remove operations, and amortized O(1) insertion.
This HashMap is a Go port of Rust's HashMap, which itself is based on Google's SwissTable algorithm. It offers high performance and resistance against HashDoS attacks.

## Overview

The HashMap struct provides a generic key-value storage solution with the following key features:

- Fast lookups using SIMD-like techniques
- Quadratic probing for collision resolution
- Amortized O(1) insertion time
- O(1) lookup and removal time
- Automatic resizing to maintain performance

## Usage

```go
import "github.com/ielm/neostd/collections/maps"

// Create a new HashMap
hm := maps.NewHashMap[string, int](collections.GenericComparator[string]())

// Insert a key-value pair
oldValue, existed := hm.Put("key", 42)

// Retrieve a value
value, found := hm.Get("key")

// Remove a key-value pair
removedValue, existed := hm.Remove("key")
```

## Key Features

### Thread Safety

The HashMap is thread-safe, using a read-write mutex to protect concurrent access.

### Custom Hashing

By default, the HashMap uses SipHash for hashing keys. However, you can provide a custom hasher for specific use cases:

```go
customHasher := &MyCustomHasher{}
hm := maps.NewHashMapWithHasher[string, int](collections.GenericComparator[string](), customHasher)
```

### Dynamic Resizing

The HashMap automatically resizes when the load factor exceeds 0.875, ensuring consistent performance as the number of elements grows.

### SIMD-like Lookup

The implementation uses SIMD-like techniques to perform fast parallel comparisons during lookups, significantly improving performance.

## Methods

### Core Operations

- `Put(key K, value V) (V, bool)`: Inserts a key-value pair, returns the old value and a boolean indicating if the key existed.
- `Get(key K) (V, bool)`: Retrieves a value by key, returns the value and a boolean indicating if the key was found.
- `Remove(key K) (V, bool)`: Removes a key-value pair, returns the removed value and a boolean indicating if the key existed.

### Utility Methods

- `Clear()`: Removes all key-value pairs from the HashMap.
- `Size() int`: Returns the number of key-value pairs in the HashMap.
- `IsEmpty() bool`: Returns true if the HashMap contains no key-value pairs.
- `Keys() []K`: Returns a slice containing all the keys in the HashMap.
- `Values() []V`: Returns a slice containing all the values in the HashMap.
- `ForEach(f func(K, V))`: Applies the given function to each key-value pair in the HashMap.
- `ContainsKey(key K) bool`: Checks if the given key exists in the HashMap.

### Advanced Features

- `SetComparator(comp comp.Comparator[K])`: Sets a custom comparator for keys, useful for complex key types or specific comparison logic.

## Implementation Details

The HashMap uses a combination of control bytes and entries to manage its internal state:

- `ctrl []byte`: Control bytes for fast SIMD-like lookups
- `entries []entry[K, V]`: Actual key-value pairs

The implementation uses quadratic probing for collision resolution and employs SIMD-like techniques for fast group matching during lookups.

### Performance Considerations

The default load factor is set to 0.875, which provides a good balance between space efficiency and performance.
The HashMap uses quadratic probing with a maximum probe distance of 128 to handle collisions efficiently.
SIMD-like techniques are used to perform fast parallel comparisons during lookups, significantly improving performance.

### References

This implementation is based on the following sources:

1. [Rust's HashMap documentation](https://doc.rust-lang.org/std/collections/struct.HashMap.html)
2. [Google's SwissTable algorithm](https://abseil.io/about/design/swisstables)

For more details on the underlying algorithm and its performance characteristics, refer to the [CppCon talk on SwissTable](https://www.youtube.com/watch?v=ncHmEUmJZf4).
