package http_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON_WritesStatusAndBody(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteJSON(rec, 201, map[string]string{"id": "abc"})

	assert.Equal(t, 201, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"id":"abc"}`, rec.Body.String())
}

func TestWriteError_NotFound(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, apperrors.NotFound("clinic abc not found"))

	assert.Equal(t, 404, rec.Code)
	var body map[string]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "NOT_FOUND", body["error"]["code"])
}

func TestWriteError_Validation(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, apperrors.Validation("invalid document", map[string]string{"document": "bad"}))

	assert.Equal(t, 422, rec.Code)
}

func TestWriteError_Conflict(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, apperrors.Conflict("already exists"))

	assert.Equal(t, 409, rec.Code)
}

func TestWriteError_UnknownErrorMapsToInternal(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, assert.AnError)

	assert.Equal(t, 500, rec.Code)
	var body map[string]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "INTERNAL_ERROR", body["error"]["code"])
}

func TestWriteError_NilErrorMapsToInternal(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, nil)

	assert.Equal(t, 500, rec.Code)
}
