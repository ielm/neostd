package res

import (
	"fmt"

	"github.com/ielm/neostd/errors"
)

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

// NewResult creates a new Result based on the given value and error.
// If err is nil, it returns an Ok Result, otherwise it returns an Err Result.
func NewResult[T any](value T, err error) Result[T] {
	if err == nil {
		return Ok(value)
	}
	return Err[T](err)
}

// Try executes the given function and returns a Result.
// If the function panics, it returns an Err Result with the panic value as the error.
func Try[T any](f func() T) (result Result[T]) {
	defer func() {
		if r := recover(); r != nil {
			result = Err[T](fmt.Errorf("%v", r))
		}
	}()
	return Ok(f())
}

// TryWithError executes the given function and returns a Result.
// If the function returns an error, it returns an Err Result with that error.
func TryWithError[T any](f func() (T, error)) Result[T] {
	value, err := f()
	return NewResult(value, err)
}

// Flatten converts a Result[Result[T]] to Result[T].
func Flatten[T any](r Result[Result[T]]) Result[T] {
	if r.IsOk() {
		return r.Unwrap()
	}
	return Err[T](r.UnwrapErr())
}

// Transpose converts a Result[Option[T]] to Option[Result[T]].
func Transpose[T any](r Result[Option[T]]) Option[Result[T]] {
	if r.IsErr() {
		return Some(Err[T](r.UnwrapErr()))
	}
	opt := r.Unwrap()
	if opt.IsNone() {
		return None[Result[T]]()
	}
	return Some(Ok(opt.Unwrap()))
}

// Collect applies a function that returns a Result to a slice of values,
// returning a Result containing a slice of successfully processed values.
func Collect[T, U any](values []T, f func(T) Result[U]) Result[[]U] {
	result := make([]U, 0, len(values))
	for _, v := range values {
		r := f(v)
		if r.IsErr() {
			return Err[[]U](r.UnwrapErr())
		}
		result = append(result, r.Unwrap())
	}
	return Ok(result)
}

// Partition separates a slice of Results into a slice of Ok values and a slice of Err values.
func Partition[T any](results []Result[T]) ([]T, []error) {
	ok := make([]T, 0, len(results))
	errs := make([]error, 0)
	for _, r := range results {
		if r.IsOk() {
			ok = append(ok, r.Unwrap())
		} else {
			errs = append(errs, r.UnwrapErr())
		}
	}
	return ok, errs
}

// Zip combines two Results into a single Result containing a pair of values.
func Zip[T, U any](r1 Result[T], r2 Result[U]) Result[struct {
	First  T
	Second U
}] {
	if r1.IsErr() {
		return Err[struct {
			First  T
			Second U
		}](r1.UnwrapErr())
	}
	if r2.IsErr() {
		return Err[struct {
			First  T
			Second U
		}](r2.UnwrapErr())
	}
	return Ok(struct {
		First  T
		Second U
	}{
		First:  r1.Unwrap(),
		Second: r2.Unwrap(),
	})
}

// FromError creates a Result from an error.
// If the error is nil, it returns an Ok Result with the zero value of T.
// Otherwise, it returns an Err Result with the given error.
func FromError[T any](err error) Result[T] {
	if err == nil {
		var zero T
		return Ok(zero)
	}
	return Err[T](err)
}

// AsError converts a Result to an error.
// If the Result is Ok, it returns nil.
// If the Result is Err, it returns the contained error.
func AsError[T any](r Result[T]) error {
	if r.IsOk() {
		return nil
	}
	return r.UnwrapErr()
}

// WrapError wraps an existing error with additional context.
func WrapError[T any](r Result[T], msg string) Result[T] {
	if r.IsErr() {
		return Err[T](fmt.Errorf("%s: %w", msg, r.UnwrapErr()))
	}
	return r
}

// UnwrapOrDefault returns the contained Ok value or the zero value of T.
func UnwrapOrDefault[T any](r Result[T]) T {
	if r.IsOk() {
		return r.Unwrap()
	}
	var zero T
	return zero
}

// Ensure Result implements the error interface
var _ error = (*Result[any])(nil)

// Error implements the error interface for Result.
func (r Result[T]) Error() string {
	if r.IsOk() {
		return "<Ok>"
	}
	return fmt.Sprintf("<Err: %v>", r.UnwrapErr())
}

// IsErrorCode checks if the Result contains an error with the specified error code.
func (r Result[T]) IsErrorCode(code errors.ErrorCode) bool {
	if r.IsErr() {
		if err, ok := r.UnwrapErr().(*errors.Error); ok {
			return err.Code == code
		}
	}
	return false
}
