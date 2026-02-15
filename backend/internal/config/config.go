package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                string
	Port               string
	DatabaseURL        string
	CORSAllowedOrigins []string
}

func Load() *Config {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	if env == "local" {
		// Non-fatal: local development can still run with exported env vars.
		_ = godotenv.Load(".env.local")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	corsAllowedOrigins := parseCSVEnv("CORS_ALLOWED_ORIGINS")
	return &Config{
		Env:                env,
		Port:               port,
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		CORSAllowedOrigins: corsAllowedOrigins,
	}
}

func parseCSVEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
