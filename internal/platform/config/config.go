package config

import "os"

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port        string
	OpenAPIPath string
}

// Load reads configuration from environment variables, falling back to
// sensible defaults when a variable is unset or empty.
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
