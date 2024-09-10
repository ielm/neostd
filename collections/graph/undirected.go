package graph

import (
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/errors"
)

// UGraph represents an undirected graph
type UGraph[V comparable, E any] struct {
	*baseGraph[V, E]
}

// NewUGraph creates a new undirected graph
func NewUGraph[V comparable, E any](comparator comp.Comparator[V]) *UGraph[V, E] {
	return &UGraph[V, E]{
		baseGraph: newBaseGraph[V, E](comparator),
	}
}

// AddEdge adds an undirected edge to the graph
func (g *UGraph[V, E]) AddEdge(source, destination V, weight E) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sourceEdges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New(errors.ErrNotFound, "source vertex not found")
	}

	destEdges, exists := g.vertices.Get(destination)
	if !exists {
		return errors.New(errors.ErrNotFound, "destination vertex not found")
	}

	sourceEdges.Put(destination, weight)
	destEdges.Put(source, weight)
	g.edgeCount++
	return nil
}

// RemoveEdge removes an undirected edge from the graph
func (g *UGraph[V, E]) RemoveEdge(source, destination V) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sourceEdges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New(errors.ErrNotFound, "source vertex not found")
	}

	destEdges, exists := g.vertices.Get(destination)
	if !exists {
		return errors.New(errors.ErrNotFound, "destination vertex not found")
	}

	if _, exists := sourceEdges.Remove(destination); exists {
		destEdges.Remove(source)
		g.edgeCount--
		return nil
	}
	return errors.New(errors.ErrNotFound, "edge not found")
}

// GetEdge returns the edge between two vertices
func (g *UGraph[V, E]) GetEdge(source, destination V) (E, bool) {
	return g.GetWeight(source, destination)
}

// GetEdges returns all edges from a vertex
func (g *UGraph[V, E]) GetEdges(vertex V) []Edge[V, E] {
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
var _ Graph[string, int] = (*UGraph[string, int])(nil)
