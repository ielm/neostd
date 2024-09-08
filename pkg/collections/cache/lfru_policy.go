package cache

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
