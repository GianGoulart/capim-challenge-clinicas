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

// WriteJSON escreve um corpo codificado em JSON com o código de status HTTP informado e
// o header Content-Type apropriado.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// WriteError mapeia um erro de domain/application para um código de status HTTP e um
// envelope de erro JSON consistente. Erros que não são *apperrors.Error
// (ex.: panics inesperados recuperados anteriormente) são tratados como internos.
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
