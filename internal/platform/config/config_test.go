package config_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_UsesDefaultsWhenUnset(t *testing.T) {
	// Usa t.Setenv (em vez de os.Unsetenv) para que o ambiente seja restaurado
	// automaticamente após este teste, mesmo que PORT/OPENAPI_PATH já estejam
	// definidas no ambiente (ex: plataformas que injetam PORT automaticamente).
	// getEnv trata uma string vazia da mesma forma que uma variável não definida.
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
