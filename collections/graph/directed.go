package graph

import (
	"github.com/ielm/neostd/collections/comp"
	"github.com/ielm/neostd/errors"
)

// DiGraph represents a directed graph
type DiGraph[V comparable, E any] struct {
	*baseGraph[V, E]
}

// NewDiGraph creates a new directed graph
func NewDiGraph[V comparable, E any](comparator comp.Comparator[V]) *DiGraph[V, E] {
	return &DiGraph[V, E]{
		baseGraph: newBaseGraph[V, E](comparator),
	}
}

// AddEdge adds a directed edge to the graph
func (g *DiGraph[V, E]) AddEdge(source, destination V, weight E) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sourceEdges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New(errors.ErrNotFound, "source vertex not found")
	}

	if _, exists := g.vertices.Get(destination); !exists {
		return errors.New(errors.ErrNotFound, "destination vertex not found")
	}

	sourceEdges.Put(destination, weight)
	g.edgeCount++
	return nil
}

// RemoveEdge removes a directed edge from the graph
func (g *DiGraph[V, E]) RemoveEdge(source, destination V) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	sourceEdges, exists := g.vertices.Get(source)
	if !exists {
		return errors.New(errors.ErrNotFound, "source vertex not found")
	}

	if _, exists := sourceEdges.Remove(destination); exists {
		g.edgeCount--
		return nil
	}
	return errors.New(errors.ErrNotFound, "edge not found")
}

// GetEdge returns the edge between two vertices
func (g *DiGraph[V, E]) GetEdge(source, destination V) (E, bool) {
	return g.GetWeight(source, destination)
}

// GetEdges returns all edges from a vertex
func (g *DiGraph[V, E]) GetEdges(vertex V) []Edge[V, E] {
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

// Ensure DirectedGraph implements the Graph interface
var _ Graph[string, int] = (*DiGraph[string, int])(nil)
