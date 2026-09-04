package config

import "os"

// Config armazena a configuração de runtime carregada a partir de variáveis de ambiente.
type Config struct {
	Port        string
	OpenAPIPath string
}

// Load lê a configuração a partir de variáveis de ambiente, usando
// defaults sensatos quando uma variável não está definida ou está vazia.
func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		OpenAPIPath: getEnv("OPENAPI_PATH", "api/openapi.yaml"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
