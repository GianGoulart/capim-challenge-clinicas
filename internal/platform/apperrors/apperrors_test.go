package apperrors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	err := apperrors.NotFound("clinic abc not found")

	assert.Equal(t, apperrors.KindNotFound, err.Kind)
	assert.Equal(t, "clinic abc not found", err.Message)
	assert.Equal(t, "NOT_FOUND: clinic abc not found", err.Error())
}

func TestValidation(t *testing.T) {
	err := apperrors.Validation("invalid document", map[string]string{"document": "bad check digit"})

	assert.Equal(t, apperrors.KindValidation, err.Kind)
	assert.Equal(t, map[string]string{"document": "bad check digit"}, err.Details)
}

func TestConflict(t *testing.T) {
	err := apperrors.Conflict("already exists")
	assert.Equal(t, apperrors.KindConflict, err.Kind)
}

func TestInternal(t *testing.T) {
	err := apperrors.Internal("boom")
	assert.Equal(t, apperrors.KindInternal, err.Kind)
}

func TestIs_MatchesWrappedError(t *testing.T) {
	base := apperrors.NotFound("clinic abc not found")
	wrapped := fmt.Errorf("loading clinic: %w", base)

	assert.True(t, apperrors.Is(wrapped, apperrors.KindNotFound))
	assert.False(t, apperrors.Is(wrapped, apperrors.KindConflict))
}

func TestIs_ReturnsFalseForPlainErrors(t *testing.T) {
	assert.False(t, apperrors.Is(errors.New("plain error"), apperrors.KindNotFound))
}
