package config_test

import (
	"os"
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_UsesDefaultsWhenUnset(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("OPENAPI_PATH")

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
