package result

type Result[T any] struct {
	Value T
	Error error
}

type Option[T any] struct {
	Value T
	Valid bool
}
