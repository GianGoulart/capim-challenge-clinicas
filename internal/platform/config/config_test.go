package config_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_UsesDefaultsWhenUnset(t *testing.T) {
	// Use t.Setenv (not os.Unsetenv) so the environment is automatically
	// restored after this test, even if PORT/OPENAPI_PATH happen to be set
	// in the ambient environment (e.g. platforms that auto-inject PORT).
	// getEnv treats an empty string the same as unset.
	t.Setenv("PORT", "")
	t.Setenv("OPENAPI_PATH", "")

	cfg := config.Load()

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "api/openapi.yaml", cfg.OpenAPIPath)
}

func TestLoad_UsesEnvironmentWhenSet(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("OPENAPI_PATH", "/custom/openapi.yaml")

	cfg := config.Load()

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "/custom/openapi.yaml", cfg.OpenAPIPath)
}
