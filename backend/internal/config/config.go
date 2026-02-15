package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
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
	return &Config{
		Env:         env,
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
