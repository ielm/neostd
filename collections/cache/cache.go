package cache

import (
	"sync"
	"time"

	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/collections/maps"
)

// Item represents a cache item
type Item[K any] struct {
	key        K
	value      interface{}
	frequency  int
	lastAccess time.Time
}

// OrderPolicy defines the interface for cache ordering policies
type OrderPolicy[K any] interface {
	Add(item *Item[K])
	Remove(item *Item[K])
	Update(item *Item[K])
	Evict() *Item[K]
}

// Cache represents the main cache structure
type Cache[K any] struct {
	capacity   int
	items      *maps.HashMap[K, *Item[K]]
	policy     OrderPolicy[K]
	mutex      sync.RWMutex
	comparator comp.Comparator[K]
}

// NewCache creates a new cache with the given capacity and order policy
// The comparator is used to compare keys in the cache, it's used by the underlying map
// to find the item in O(1) time
func NewCache[K any](capacity int, policy OrderPolicy[K], comparator comp.Comparator[K]) *Cache[K] {
	return &Cache[K]{
		capacity:   capacity,
		items:      maps.NewHashMap[K, *Item[K]](comparator),
		policy:     policy,
		comparator: comparator,
	}
}

// Set adds or updates an item in the cache
func (c *Cache[K]) Set(key K, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.items.Get(key); ok {
		item.value = value
		item.lastAccess = time.Now()
		item.frequency++
		c.policy.Update(item)
	} else {
		if c.items.Size() >= c.capacity {
			c.evict()
		}
		item := &Item[K]{
			key:        key,
			value:      value,
			frequency:  1,
			lastAccess: time.Now(),
		}
		c.policy.Add(item)
		c.items.Put(key, item)
	}
}

// Get retrieves an item from the cache
func (c *Cache[K]) Get(key K) (interface{}, bool) {
	c.mutex.RLock()
	item, ok := c.items.Get(key)
	c.mutex.RUnlock()

	if !ok {
		return nil, false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	item.lastAccess = time.Now()
	item.frequency++
	c.policy.Update(item)
	return item.value, true
}

// evict removes the item selected by the order policy
func (c *Cache[K]) evict() {
	if item := c.policy.Evict(); item != nil {
		c.items.Remove(item.key)
	}
}

// Remove removes an item from the cache
func (c *Cache[K]) Remove(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.items.Get(key); ok {
		c.policy.Remove(item)
		c.items.Remove(key)
	}
}

// Clear removes all items from the cache
func (c *Cache[K]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = maps.NewHashMap[K, *Item[K]](c.comparator)
	c.policy = c.createNewPolicy()
}

// Size returns the number of items in the cache
func (c *Cache[K]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.items.Size()
}

// Update the createNewPolicy method to include the new policies
func (c *Cache[K]) createNewPolicy() OrderPolicy[K] {
	switch c.policy.(type) {
	case *LRUPolicy[K]:
		return NewLRUPolicy[K]()
	case *LFUPolicy[K]:
		return NewLFUPolicy[K]()
	case *LFRUPolicy[K]:
		return NewLFRUPolicy[K]()
	default:
		panic("Unknown policy type")
	}
}
