package config

import "os"

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
