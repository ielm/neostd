package result

import "fmt"

type Result[T any] struct {
	value T
	err   error
	isOk  bool
}

// Creates a new Result with a success value
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, isOk: true}
}

// Creates a new Result with an error value
func Err[T any](err error) Result[T] {
	return Result[T]{err: err, isOk: false}
}

// IsOk returns true if the Result is Ok.
func (r Result[T]) IsOk() bool {
	return r.isOk
}

// IsErr returns true if the Result is Err.
func (r Result[T]) IsErr() bool {
	return !r.isOk
}

// Unwrap returns the contained Ok value if the Result is Ok, otherwise panics.
func (r Result[T]) Unwrap() T {
	if !r.isOk {
		panic(fmt.Sprintf("called Result.Unwrap() on an Err value: %v", r.err))
	}
	return r.value
}

// UnwrapOr returns the contained Ok value or a provided default.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.isOk {
		return r.value
	}
	return defaultValue
}

// UnwrapOrElse returns the contained Ok value or computes it from a closure.
func (r Result[T]) UnwrapOrElse(f func() T) T {
	if r.isOk {
		return r.value
	}
	return f()
}

// Expect returns the contained Ok value or panics with a custom error message.
func (r Result[T]) Expect(msg string) T {
	if !r.isOk {
		panic(fmt.Sprintf("%s: %v", msg, r.err))
	}
	return r.value
}

// UnwrapErr returns the contained Err value if the Result is Err, otherwise panics.
func (r Result[T]) UnwrapErr() error {
	if r.isOk {
		panic("called Result.UnwrapErr() on an Ok value")
	}
	return r.err
}

// Map applies a function to the contained value (if Ok), or returns the original error (if Err).
func (r Result[T]) Map(f func(T) T) Result[T] {
	if r.isOk {
		return Ok(f(r.value))
	}
	return r
}

// MapErr applies a function to the contained error (if Err), or returns the original value (if Ok).
func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if !r.isOk {
		return Err[T](f(r.err))
	}
	return r
}

// And returns res if the Result is Ok, otherwise returns the Err value of self.
func (r Result[T]) And(res Result[T]) Result[T] {
	if r.isOk {
		return res
	}
	return r
}

// AndThen calls op if the Result is Ok, otherwise returns the Err value of self.
func (r Result[T]) AndThen(op func(T) Result[T]) Result[T] {
	if r.isOk {
		return op(r.value)
	}
	return r
}

// Or returns res if the Result is Err, otherwise returns the Ok value of self.
func (r Result[T]) Or(res Result[T]) Result[T] {
	if r.isOk {
		return r
	}
	return res
}

// OrElse calls op if the Result is Err, otherwise returns the Ok value of self.
func (r Result[T]) OrElse(op func(error) Result[T]) Result[T] {
	if r.isOk {
		return r
	}
	return op(r.err)
}

// Match applies the appropriate function based on the Result variant.
func (r Result[T]) Match(okFn func(T), errFn func(error)) {
	if r.isOk {
		okFn(r.value)
	} else {
		errFn(r.err)
	}
}

// ToOption converts the Result to an Option type.
func (r Result[T]) ToOption() Option[T] {
	if r.isOk {
		return Some(r.value)
	}
	return None[T]()
}
