package errors

import (
	"fmt"
	"runtime"
	"strings"
)

type ErrorCode int

const (
	ErrInvalidArgument ErrorCode = iota
	ErrConstructionFailed
	ErrOutOfBounds
	ErrNotFound
	ErrNotImplemented
	ErrUnwrapOnErr
	ErrInternal
	// ...
)

type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
	Stack   []uintptr
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) StackTrace() string {
	var sb strings.Builder
	frames := runtime.CallersFrames(e.Stack)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&sb, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return sb.String()
}

func New(code ErrorCode, message string) *Error {
	return NewWithCause(code, message, nil)
}

func NewWithCause(code ErrorCode, message string, cause error) *Error {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
		Stack:   pcs[:n],
	}
}

func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return NewWithCause(e.Code, fmt.Sprintf("%s: %s", message, e.Message), e.Cause)
	}
	return NewWithCause(ErrInvalidArgument, message, err)
}

func Is(err, target error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func As(err error, target interface{}) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	t, ok := target.(**Error)
	if !ok {
		return false
	}
	*t = e
	return true
}
