package cache

import (
	"sync"
	"time"

	"github.com/ielm/neostd/pkg/collections"
	"github.com/ielm/neostd/pkg/collections/list"
	"github.com/ielm/neostd/pkg/collections/maps"
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
	comparator collections.Comparator[K]
}

// NewCache creates a new cache with the given capacity and order policy
func NewCache[K any](capacity int, policy OrderPolicy[K], comparator collections.Comparator[K]) *Cache[K] {
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

// LRUPolicy implements the Least Recently Used order policy
type LRUPolicy[K any] struct {
	list *list.LinkedList[*Item[K]]
}

func NewLRUPolicy[K any]() *LRUPolicy[K] {
	return &LRUPolicy[K]{list: list.NewLinkedList[*Item[K]]()}
}

func (p *LRUPolicy[K]) Add(item *Item[K]) {
	p.list.AddFirst(item)
}

func (p *LRUPolicy[K]) Remove(item *Item[K]) {
	p.list.Remove(item)
}

func (p *LRUPolicy[K]) Update(item *Item[K]) {
	p.Remove(item)
	p.Add(item)
}

func (p *LRUPolicy[K]) Evict() *Item[K] {
	if p.list.IsEmpty() {
		return nil
	}
	item, _ := p.list.RemoveLast()
	return item
}

// LFUPolicy implements the Least Frequently Used order policy
type LFUPolicy[K any] struct {
	frequencies *list.LinkedList[*FrequencyNode[K]]
	items       map[*Item[K]]*list.Node[*Item[K]]
	freqMap     map[int]*FrequencyNode[K]
	minFreq     int
}

type FrequencyNode[K any] struct {
	frequency int
	items     *list.LinkedList[*Item[K]]
}

func NewLFUPolicy[K any]() *LFUPolicy[K] {
	return &LFUPolicy[K]{
		frequencies: list.NewLinkedList[*FrequencyNode[K]](),
		items:       make(map[*Item[K]]*list.Node[*Item[K]]),
		freqMap:     make(map[int]*FrequencyNode[K]),
		minFreq:     0,
	}
}

func (p *LFUPolicy[K]) Add(item *Item[K]) {
	freqNode := p.getOrCreateFrequencyNode(1)
	itemNode := freqNode.items.AddLast(item)
	p.items[item] = itemNode
	p.minFreq = 1
}

func (p *LFUPolicy[K]) Remove(item *Item[K]) {
	if itemNode, ok := p.items[item]; ok {
		freqNode := p.freqMap[item.frequency]
		freqNode.items.RemoveNode(itemNode)
		delete(p.items, item)

		if freqNode.items.IsEmpty() {
			p.removeFrequencyNode(freqNode)
		}
	}
}

func (p *LFUPolicy[K]) Update(item *Item[K]) {
	p.Remove(item)
	item.frequency++
	freqNode := p.getOrCreateFrequencyNode(item.frequency)
	itemNode := freqNode.items.AddLast(item)
	p.items[item] = itemNode

	if item.frequency-1 == p.minFreq && p.freqMap[p.minFreq] == nil {
		p.minFreq = item.frequency
	}
}

func (p *LFUPolicy[K]) Evict() *Item[K] {
	if p.minFreq > 0 && p.freqMap[p.minFreq] != nil {
		freqNode := p.freqMap[p.minFreq]
		itemNode := freqNode.items.First()
		if itemNode != nil {
			item := itemNode.Value()
			p.Remove(item)
			return item
		}
	}
	return nil
}

func (p *LFUPolicy[K]) getOrCreateFrequencyNode(freq int) *FrequencyNode[K] {
	if freqNode, ok := p.freqMap[freq]; ok {
		return freqNode
	}

	freqNode := &FrequencyNode[K]{
		frequency: freq,
		items:     list.NewLinkedList[*Item[K]](),
	}
	p.freqMap[freq] = freqNode
	// Insert the new frequency node in the correct position
	var insertAfter *list.Node[*FrequencyNode[K]]
	for node := p.frequencies.First(); node != nil; node = node.Next() {
		if node.Value().frequency > freq {
			break
		}
		insertAfter = node
	}

	if insertAfter == nil {
		p.frequencies.AddFirst(freqNode)
	} else {
		p.frequencies.AddAfter(insertAfter, freqNode)
	}

	return freqNode
}

func (p *LFUPolicy[K]) removeFrequencyNode(freqNode *FrequencyNode[K]) {
	p.frequencies.Remove(freqNode)
	delete(p.freqMap, freqNode.frequency)
}

// LFRUPolicy implements the Least Frequently/Recently Used order policy
type LFRUPolicy[K any] struct {
	lfu *LFUPolicy[K]
	lru *LRUPolicy[K]
}

func NewLFRUPolicy[K any]() *LFRUPolicy[K] {
	return &LFRUPolicy[K]{
		lfu: NewLFUPolicy[K](),
		lru: NewLRUPolicy[K](),
	}
}

func (p *LFRUPolicy[K]) Add(item *Item[K]) {
	p.lfu.Add(item)
	p.lru.Add(item)
}

func (p *LFRUPolicy[K]) Remove(item *Item[K]) {
	p.lfu.Remove(item)
	p.lru.Remove(item)
}

func (p *LFRUPolicy[K]) Update(item *Item[K]) {
	p.lfu.Update(item)
	p.lru.Update(item)
}

func (p *LFRUPolicy[K]) Evict() *Item[K] {
	lfuItem := p.lfu.Evict()
	lruItem := p.lru.Evict()

	if lfuItem == lruItem {
		return lfuItem
	}

	if lfuItem.frequency < lruItem.frequency {
		p.lru.Add(lruItem)
		return lfuItem
	}

	p.lfu.Add(lfuItem)
	return lruItem
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
