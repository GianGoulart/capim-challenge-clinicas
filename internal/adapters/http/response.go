package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

// WriteJSON writes a JSON-encoded body with the given HTTP status code and
// the appropriate Content-Type header.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// WriteError maps a domain/application error to an HTTP status code and a
// consistent JSON error envelope. Errors that are not *apperrors.Error
// (e.g. unexpected panics recovered upstream) are treated as internal.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *apperrors.Error
	status := http.StatusInternalServerError
	code := string(apperrors.KindInternal)
	message := "internal server error"
	var details map[string]string

	if errors.As(err, &appErr) {
		code = string(appErr.Kind)
		message = appErr.Message
		details = appErr.Details
		switch appErr.Kind {
		case apperrors.KindNotFound:
			status = http.StatusNotFound
		case apperrors.KindValidation:
			status = http.StatusUnprocessableEntity
		case apperrors.KindConflict:
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
		}
	}

	WriteJSON(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message, Details: details}})
}
