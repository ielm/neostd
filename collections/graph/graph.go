package graph

import (
	"errors"
	"sync"

	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/collections/maps"
)

// Edge represents an edge in the graph
type Edge[V any, E any] struct {
	Source      V
	Destination V
	Weight      E
}

// Graph is the interface that both directed and undirected graphs implement
type Graph[V comparable, E any] interface {
	collections.Collection[V]
	AddEdge(source, destination V, weight E) error
	RemoveEdge(source, destination V) error
	GetEdge(source, destination V) (E, bool)
	GetEdges(vertex V) []Edge[V, E]
	GetVertices() []V
	GetNeighbors(vertex V) []V
	HasEdge(source, destination V) bool
	GetWeight(source, destination V) (E, bool)
	SetWeight(source, destination V, weight E) error
}

// baseGraph is the common implementation for both directed and undirected graphs
type baseGraph[V comparable, E any] struct {
	vertices   *maps.HashMap[V, *maps.HashMap[V, E]]
	edgeCount  int
	comparator comp.Comparator[V]
	mu         sync.RWMutex
}

// newBaseGraph creates a new base graph
func newBaseGraph[V comparable, E any](comparator comp.Comparator[V]) *baseGraph[V, E] {
	return &baseGraph[V, E]{
		vertices:   maps.NewHashMap[V, *maps.HashMap[V, E]](comparator),
		comparator: comparator,
	}
}

// Add adds a vertex to the graph
func (g *baseGraph[V, E]) Add(vertex V) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.vertices.Get(vertex); exists {
		return false
	}
	g.vertices.Put(vertex, maps.NewHashMap[V, E](g.comparator))
	return true
}

// Remove removes a vertex and all its edges from the graph
func (g *baseGraph[V, E]) Remove(vertex V) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	edges, exists := g.vertices.Get(vertex)
	if !exists {
		return false
	}

	g.edgeCount -= edges.Size()
	g.vertices.Remove(vertex)

	g.vertices.ForEach(func(v V, edges *maps.HashMap[V, E]) {
		if _, ok := edges.Remove(vertex); ok {
			g.edgeCount--
		}
	})

	return true
}

// Contains checks if the graph contains a vertex
func (g *baseGraph[V, E]) Contains(vertex V) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	_, exists := g.vertices.Get(vertex)
	return exists
}

// Size returns the number of vertices in the graph
func (g *baseGraph[V, E]) Size() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.vertices.Size()
}

// Clear removes all vertices and edges from the graph
func (g *baseGraph[V, E]) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.vertices.Clear()
	g.edgeCount = 0
}

// IsEmpty returns true if the graph has no vertices
func (g *baseGraph[V, E]) IsEmpty() bool {
	return g.Size() == 0
}

// SetComparator sets the comparator for the graph
func (g *baseGraph[V, E]) SetComparator(comp comp.Comparator[V]) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.comparator = comp
	g.vertices.SetComparator(comp)
	g.vertices.ForEach(func(_ V, edges *maps.HashMap[V, E]) {
		edges.SetComparator(comp)
	})
}

// GetVertices returns a slice of all vertices in the graph
func (g *baseGraph[V, E]) GetVertices() []V {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.vertices.Keys()
}

// GetNeighbors returns a slice of all neighbors of a vertex
func (g *baseGraph[V, E]) GetNeighbors(vertex V) []V {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges, exists := g.vertices.Get(vertex)
	if !exists {
		return []V{}
	}
	return edges.Keys()
}

// HasEdge checks if an edge exists between two vertices
func (g *baseGraph[V, E]) HasEdge(source, destination V) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges, exists := g.vertices.Get(source)
	if !exists {
		return false
	}
	_, exists = edges.Get(destination)
	return exists
}

// GetWeight returns the weight of an edge between two vertices
func (g *baseGraph[V, E]) GetWeight(source, destination V) (E, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges, exists := g.vertices.Get(source)
	if !exists {
		var zero E
		return zero, false
	}
	return edges.Get(destination)
}

// SetWeight sets the weight of an edge between two vertices
func (g *baseGraph[V, E]) SetWeight(source, destination V, weight E) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	edges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New("source vertex not found")
	}
	if _, exists := edges.Get(destination); !exists {
		return errors.New("edge not found")
	}
	edges.Put(destination, weight)
	return nil
}

// Iterator returns an iterator over the vertices of the graph
func (g *baseGraph[V, E]) Iterator() collections.Iterator[V] {
	return &graphIterator[V, E]{
		graph: g,
		keys:  g.GetVertices(),
		index: 0,
	}
}

// ReverseIterator returns a reverse iterator over the vertices of the graph
func (g *baseGraph[V, E]) ReverseIterator() collections.Iterator[V] {
	keys := g.GetVertices()
	return &graphIterator[V, E]{
		graph:   g,
		keys:    keys,
		index:   len(keys) - 1,
		reverse: true,
	}
}

type graphIterator[V comparable, E any] struct {
	graph   *baseGraph[V, E]
	keys    []V
	index   int
	reverse bool
}

func (it *graphIterator[V, E]) HasNext() bool {
	if it.reverse {
		return it.index >= 0
	}
	return it.index < len(it.keys)
}

func (it *graphIterator[V, E]) Next() V {
	if !it.HasNext() {
		panic("no more elements")
	}
	vertex := it.keys[it.index]
	if it.reverse {
		it.index--
	} else {
		it.index++
	}
	return vertex
}

// Ensure baseGraph implements the Collection interface
var _ collections.Collection[string] = (*baseGraph[string, int])(nil)
