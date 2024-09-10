package result

type Result[T any] struct {
	value T
	err   error
	isOk  bool
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, isOk: true}
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err, isOk: false}
}
