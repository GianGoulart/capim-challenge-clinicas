package config

import "os"

type Config struct {
	Port        string
	OpenAPIPath string
}

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
