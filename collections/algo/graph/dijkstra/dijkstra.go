package dijkstra

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/graph"
	"github.com/ielm/neostd/collections/heap"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/res"
)

// DijkstraResult represents the result of Dijkstra's algorithm
type DijkstraResult[V comparable, E any] struct {
	Distances    map[V]E
	Predecessors map[V]V
}

// Dijkstra performs Dijkstra's algorithm on the given graph
func Dijkstra[V comparable, E any](
	g graph.Graph[V, E],
	start V,
	less func(E, E) bool,
	zero E,
	add func(E, E) E,
) res.Result[DijkstraResult[V, E]] {
	distances := make(map[V]E)
	predecessors := make(map[V]V)
	visited := make(map[V]bool)

	// Initialize distances
	for _, v := range g.GetVertices() {
		distances[v] = zero
	}
	distances[start] = zero

	// Create a min-heap priority queue
	pq := heap.NewMinBinaryHeap(func(a, b collections.Pair[V, E]) int {
		if less(a.Value, b.Value) {
			return -1
		}
		if less(b.Value, a.Value) {
			return 1
		}
		return 0
	})

	pq.Push(collections.Pair[V, E]{Key: start, Value: zero})

	for !pq.IsEmpty() {
		current := pq.Pop().Unwrap()
		currentVertex := current.Key
		currentDist := current.Value

		if visited[currentVertex] {
			continue
		}
		visited[currentVertex] = true

		if less(currentDist, distances[currentVertex]) {
			distances[currentVertex] = currentDist
		}

		for _, neighbor := range g.GetNeighbors(currentVertex) {
			if visited[neighbor] {
				continue
			}

			weight, ok := g.GetWeight(currentVertex, neighbor)
			if !ok {
				return res.Err[DijkstraResult[V, E]](errors.New(errors.ErrInternal, "edge weight not found"))
			}

			newDist := add(currentDist, weight)
			if less(newDist, distances[neighbor]) {
				distances[neighbor] = newDist
				predecessors[neighbor] = currentVertex
				pq.Push(collections.Pair[V, E]{Key: neighbor, Value: newDist})
			}
		}
	}

	return res.Ok(DijkstraResult[V, E]{
		Distances:    distances,
		Predecessors: predecessors,
	})
}

// ShortestPath reconstructs the shortest path from the start to the end vertex
func ShortestPath[V comparable, E any](result DijkstraResult[V, E], end V) res.Result[[]V] {
	path := []V{end}
	current := end

	for {
		prev, ok := result.Predecessors[current]
		if !ok {
			break
		}
		path = append([]V{prev}, path...)
		current = prev
	}

	if len(path) == 1 && path[0] != end {
		return res.Err[[]V](errors.New(errors.ErrNotFound, "no path found"))
	}

	return res.Ok(path)
}
