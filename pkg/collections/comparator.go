package collections

import "golang.org/x/exp/constraints"

// Comparator is a function type that compares two values
// It returns a negative value if a < b, zero if a == b, and a positive value if a > b
type Comparator[T any] func(a, b T) int

// GenericComparator returns a Comparator for any ordered type
func GenericComparator[T constraints.Ordered]() Comparator[T] {
	return func(a, b T) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}
}

// ReverseComparator returns a reversed Comparator
func ReverseComparator[T any](cmp Comparator[T]) Comparator[T] {
	return func(a, b T) int {
		return -cmp(a, b)
	}
}

// ChainComparators chains multiple Comparators
func ChainComparators[T any](comparators ...Comparator[T]) Comparator[T] {
	return func(a, b T) int {
		for _, cmp := range comparators {
			if result := cmp(a, b); result != 0 {
				return result
			}
		}
		return 0
	}
}

// Min returns the minimum of two values
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two values
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
