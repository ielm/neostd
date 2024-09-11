package common

type Countable interface {
	Size() int
	IsEmpty() bool
}

type Clearable interface {
	Clear()
}
