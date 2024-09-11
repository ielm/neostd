package astar

import (
	"github.com/ielm/neostd/collections"
	"github.com/ielm/neostd/collections/graph"
	"github.com/ielm/neostd/collections/heap"
	"github.com/ielm/neostd/collections/maps"
	"github.com/ielm/neostd/errors"
	"github.com/ielm/neostd/res"
)

// AStarResult represents the result of A* algorithm
type AStarResult[V comparable, E any] struct {
	Path     []V
	Cost     E
	Explored int
}

// AStar performs A* search algorithm on the given graph
func AStar[V comparable, E any](
	g graph.Graph[V, E],
	start, goal V,
	heuristic func(V) E,
	less func(E, E) bool,
	zero E,
	add func(E, E) E,
) res.Result[AStarResult[V, E]] {
	openSet := heap.NewMinBinaryHeap(func(a, b collections.Pair[V, E]) int {
		if less(a.Value, b.Value) {
			return -1
		}
		if less(b.Value, a.Value) {
			return 1
		}
		return 0
	})
	openSet.Push(collections.Pair[V, E]{Key: start, Value: zero})

	cameFrom := maps.NewHashMap[V, V](g.Comparator()).Unwrap()
	gScore := maps.NewHashMap[V, E](g.Comparator()).Unwrap()
	gScore.Put(start, zero)

	fScore := maps.NewHashMap[V, E](g.Comparator()).Unwrap()
	fScore.Put(start, heuristic(start))

	explored := 0

	for !openSet.IsEmpty() {
		current := openSet.Pop().Unwrap().Key
		explored++

		if g.Comparator()(current, goal) == 0 {
			path := reconstructPath(cameFrom, current)
			cost, _ := gScore.Get(current)
			return res.Ok(AStarResult[V, E]{
				Path:     path,
				Cost:     cost,
				Explored: explored,
			})
		}

		for _, neighbor := range g.GetNeighbors(current) {
			weight, ok := g.GetWeight(current, neighbor)
			if !ok {
				return res.Err[AStarResult[V, E]](errors.New(errors.ErrInternal, "edge weight not found"))
			}
			currentGScore, _ := gScore.Get(current)

			tentativeGScore := add(currentGScore, weight)

			neighborGScore, exists := gScore.Get(neighbor)
			if !exists || less(tentativeGScore, neighborGScore) {
				cameFrom.Put(neighbor, current)
				gScore.Put(neighbor, tentativeGScore)
				newFScore := add(tentativeGScore, heuristic(neighbor))
				fScore.Put(neighbor, newFScore)

				if !openSet.Contains(collections.Pair[V, E]{Key: neighbor}) {
					openSet.Push(collections.Pair[V, E]{Key: neighbor, Value: newFScore})
				}
			}
		}
	}

	return res.Err[AStarResult[V, E]](errors.New(errors.ErrNotFound, "no path found"))
}

// reconstructPath reconstructs the path from start to goal
func reconstructPath[V comparable](cameFrom *maps.HashMap[V, V], current V) []V {
	path := []V{current}
	for {
		prev, exists := cameFrom.Get(current)
		if !exists {
			break
		}
		path = append([]V{prev}, path...)
		current = prev
	}
	return path
}

// AStarWithOptions performs A* search with additional options
func AStarWithOptions[V comparable, E any](
	g graph.Graph[V, E],
	start, goal V,
	heuristic func(V) E,
	less func(E, E) bool,
	zero E,
	add func(E, E) E,
	options ...AStarOption[V, E],
) res.Result[AStarResult[V, E]] {
	config := defaultAStarConfig[V, E]()
	for _, option := range options {
		option(&config)
	}

	return astarWithConfig(g, start, goal, heuristic, less, zero, add, config)
}

// AStarOption represents an option for configuring A* search
type AStarOption[V comparable, E any] func(*astarConfig[V, E])

type astarConfig[V comparable, E any] struct {
	maxIterations int
	earlyExit     func(V, E) bool
	onExplore     func(V, E)
}

func defaultAStarConfig[V comparable, E any]() astarConfig[V, E] {
	return astarConfig[V, E]{
		maxIterations: -1,
		earlyExit:     func(V, E) bool { return false },
		onExplore:     func(V, E) {},
	}
}

// WithMaxIterations sets the maximum number of iterations for A* search
func WithMaxIterations[V comparable, E any](maxIterations int) AStarOption[V, E] {
	return func(c *astarConfig[V, E]) {
		c.maxIterations = maxIterations
	}
}

// WithEarlyExit sets an early exit condition for A* search
func WithEarlyExit[V comparable, E any](earlyExit func(V, E) bool) AStarOption[V, E] {
	return func(c *astarConfig[V, E]) {
		c.earlyExit = earlyExit
	}
}

// WithOnExplore sets a callback function to be called when a node is explored
func WithOnExplore[V comparable, E any](onExplore func(V, E)) AStarOption[V, E] {
	return func(c *astarConfig[V, E]) {
		c.onExplore = onExplore
	}
}

func astarWithConfig[V comparable, E any](
	g graph.Graph[V, E],
	start, goal V,
	heuristic func(V) E,
	less func(E, E) bool,
	zero E,
	add func(E, E) E,
	config astarConfig[V, E],
) res.Result[AStarResult[V, E]] {
	openSet := heap.NewMinBinaryHeap(func(a, b collections.Pair[V, E]) int {
		if less(a.Value, b.Value) {
			return -1
		}
		if less(b.Value, a.Value) {
			return 1
		}
		return 0
	})
	openSet.Push(collections.Pair[V, E]{Key: start, Value: zero})

	cameFrom := maps.NewHashMap[V, V](g.Comparator()).Unwrap()
	gScore := maps.NewHashMap[V, E](g.Comparator()).Unwrap()
	gScore.Put(start, zero)

	fScore := maps.NewHashMap[V, E](g.Comparator()).Unwrap()
	fScore.Put(start, heuristic(start))

	explored := 0

	for !openSet.IsEmpty() && (config.maxIterations == -1 || explored < config.maxIterations) {
		current := openSet.Pop().Unwrap().Key
		currentGScore, _ := gScore.Get(current)
		explored++

		config.onExplore(current, currentGScore)

		if g.Comparator()(current, goal) == 0 || config.earlyExit(current, currentGScore) {
			path := reconstructPath(cameFrom, current)
			return res.Ok(AStarResult[V, E]{
				Path:     path,
				Cost:     currentGScore,
				Explored: explored,
			})
		}

		for _, neighbor := range g.GetNeighbors(current) {
			weight, ok := g.GetWeight(current, neighbor)
			if !ok {
				return res.Err[AStarResult[V, E]](errors.New(errors.ErrInternal, "edge weight not found"))
			}

			tentativeGScore := add(currentGScore, weight)

			neighborGScore, exists := gScore.Get(neighbor)
			if !exists || less(tentativeGScore, neighborGScore) {
				cameFrom.Put(neighbor, current)
				gScore.Put(neighbor, tentativeGScore)
				newFScore := add(tentativeGScore, heuristic(neighbor))
				fScore.Put(neighbor, newFScore)

				if !openSet.Contains(collections.Pair[V, E]{Key: neighbor}) {
					openSet.Push(collections.Pair[V, E]{Key: neighbor, Value: newFScore})
				}
			}
		}
	}

	return res.Err[AStarResult[V, E]](errors.New(errors.ErrNotFound, "no path found"))
}
