package graph

import (
	"errors"

	"github.com/ielm/neostd/collections/comp"
)

// UndirectedGraph represents an undirected graph
type UndirectedGraph[V comparable, E any] struct {
	*baseGraph[V, E]
}

// NewUndirectedGraph creates a new undirected graph
func NewUndirectedGraph[V comparable, E any](comparator comp.Comparator[V]) *UndirectedGraph[V, E] {
	return &UndirectedGraph[V, E]{
		baseGraph: newBaseGraph[V, E](comparator),
	}
}

// AddEdge adds an undirected edge to the graph
func (g *UndirectedGraph[V, E]) AddEdge(source, destination V, weight E) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sourceEdges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New("source vertex not found")
	}

	destEdges, exists := g.vertices.Get(destination)
	if !exists {
		return errors.New("destination vertex not found")
	}

	sourceEdges.Put(destination, weight)
	destEdges.Put(source, weight)
	g.edgeCount++
	return nil
}

// RemoveEdge removes an undirected edge from the graph
func (g *UndirectedGraph[V, E]) RemoveEdge(source, destination V) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sourceEdges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New("source vertex not found")
	}

	destEdges, exists := g.vertices.Get(destination)
	if !exists {
		return errors.New("destination vertex not found")
	}

	if _, exists := sourceEdges.Remove(destination); exists {
		destEdges.Remove(source)
		g.edgeCount--
		return nil
	}
	return errors.New("edge not found")
}

// GetEdge returns the edge between two vertices
func (g *UndirectedGraph[V, E]) GetEdge(source, destination V) (E, bool) {
	return g.GetWeight(source, destination)
}

// GetEdges returns all edges from a vertex
func (g *UndirectedGraph[V, E]) GetEdges(vertex V) []Edge[V, E] {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges, exists := g.vertices.Get(vertex)
	if !exists {
		return []Edge[V, E]{}
	}

	result := make([]Edge[V, E], 0, edges.Size())
	edges.ForEach(func(dest V, weight E) {
		result = append(result, Edge[V, E]{Source: vertex, Destination: dest, Weight: weight})
	})
	return result
}

// Ensure UndirectedGraph implements the Graph interface
var _ Graph[string, int] = (*UndirectedGraph[string, int])(nil)
