package apperrors

import (
	"errors"
	"fmt"
)

// Kind identifies the category of a domain-level error, used by adapters to
// map it to protocol-specific responses (e.g. HTTP status codes).
type Kind string

const (
	KindNotFound   Kind = "NOT_FOUND"
	KindValidation Kind = "VALIDATION_ERROR"
	KindConflict   Kind = "CONFLICT"
	KindInternal   Kind = "INTERNAL_ERROR"
)

// Error is the concrete error type returned by domain and application code.
// It never carries HTTP/transport-specific data — adapters translate Kind
// into whatever the protocol needs.
type Error struct {
	Kind    Kind
	Message string
	Details map[string]string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

func Validation(message string, details map[string]string) *Error {
	return &Error{Kind: KindValidation, Message: message, Details: details}
}

func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

func Internal(message string) *Error {
	return &Error{Kind: KindInternal, Message: message}
}

// Is reports whether err is (or wraps) an *Error of the given Kind.
func Is(err error, kind Kind) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Kind == kind
	}
	return false
}
