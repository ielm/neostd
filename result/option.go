package result

// Option represents an optional value.
type Option[T any] struct {
	value  T
	isSome bool
}

// Some creates a new Option with a value.
func Some[T any](value T) Option[T] {
	return Option[T]{
		value:  value,
		isSome: true,
	}
}

// None creates a new Option without a value.
func None[T any]() Option[T] {
	return Option[T]{
		isSome: false,
	}
}

// IsSome returns true if the Option is Some.
func (o Option[T]) IsSome() bool {
	return o.isSome
}

// IsNone returns true if the Option is None.
func (o Option[T]) IsNone() bool {
	return !o.isSome
}

// Unwrap returns the contained Some value, panics if the Option is None.
func (o Option[T]) Unwrap() T {
	if !o.isSome {
		panic("called Option.Unwrap() on a None value")
	}
	return o.value
}

// UnwrapOr returns the contained Some value or a provided default.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.isSome {
		return o.value
	}
	return defaultValue
}

// UnwrapOrElse returns the contained Some value or computes it from a closure.
func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.isSome {
		return o.value
	}
	return f()
}

// Map applies a function to the contained value (if Some).
func (o Option[T]) Map(f func(T) T) Option[T] {
	if o.isSome {
		return Some(f(o.value))
	}
	return o
}

// AndThen returns None if the Option is None, otherwise calls f with the wrapped value and returns the result.
func (o Option[T]) AndThen(f func(T) Option[T]) Option[T] {
	if o.isSome {
		return f(o.value)
	}
	return o
}

// Or returns the Option if it contains a value, otherwise returns optb.
func (o Option[T]) Or(optb Option[T]) Option[T] {
	if o.isSome {
		return o
	}
	return optb
}

// OrElse returns the Option if it contains a value, otherwise calls f and returns the result.
func (o Option[T]) OrElse(f func() Option[T]) Option[T] {
	if o.isSome {
		return o
	}
	return f()
}

// Match applies the appropriate function based on the Option variant.
func (o Option[T]) Match(someFn func(T), noneFn func()) {
	if o.isSome {
		someFn(o.value)
	} else {
		noneFn()
	}
}

// ToResult converts the Option to a Result type.
func (o Option[T]) ToResult(err error) Result[T] {
	if o.isSome {
		return Ok(o.value)
	}
	return Err[T](err)
}
