package set

import "github.com/ielm/neostd/internal/common"

type ProbabilisticSet[T any] interface {
	common.Countable
	common.Clearable
	Add(item T) bool
	Contains(item T) bool
}
