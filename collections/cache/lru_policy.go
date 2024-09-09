package cache

import "github.com/ielm/neostd/collections/list"

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
	p.list.MoveToFront(item)
}

func (p *LRUPolicy[K]) Evict() *Item[K] {
	if p.list.IsEmpty() {
		return nil
	}
	item, _ := p.list.RemoveLast()
	return item
}
