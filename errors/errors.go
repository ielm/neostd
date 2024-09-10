package errors

import "fmt"

type ErrorCode int

const (
	ErrInvalidArgument    ErrorCode = iota
	ErrConstructionFailed ErrorCode = iota
	ErrOutOfBounds
	ErrNotFound
	ErrNotImplemented
)

type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func New(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}
