package apperrors

import (
	"errors"
	"fmt"
)

// Kind identifica a categoria de um erro de domínio, usado pelos adapters para
// mapeá-lo para respostas específicas do protocolo (ex: HTTP status codes).
type Kind string

const (
	KindNotFound   Kind = "NOT_FOUND"
	KindValidation Kind = "VALIDATION_ERROR"
	KindConflict   Kind = "CONFLICT"
	KindInternal   Kind = "INTERNAL_ERROR"
)

// Error é o tipo concreto de erro retornado pelo código de domínio e de aplicação.
// Ele nunca carrega dados específicos de HTTP/transporte — os adapters traduzem Kind
// para o que o protocolo precisar.
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

// Is reporta se err é (ou envolve) um *Error do Kind informado.
func Is(err error, kind Kind) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Kind == kind
	}
	return false
}
