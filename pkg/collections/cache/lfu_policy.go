package cache

import "github.com/ielm/neostd/pkg/collections/list"

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
