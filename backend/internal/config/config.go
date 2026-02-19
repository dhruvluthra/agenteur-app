package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                string
	Port               string
	DatabaseURL        string
	CORSAllowedOrigins []string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	InviteBaseURL      string
	InviteTokenTTL     time.Duration
	BcryptCost         int
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

	accessTTL := parseDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	refreshTTL := parseDuration("REFRESH_TOKEN_TTL", 168*time.Hour)
	inviteTTL := parseDuration("INVITE_TOKEN_TTL", 72*time.Hour)

	inviteBaseURL := os.Getenv("INVITE_BASE_URL")
	if inviteBaseURL == "" {
		inviteBaseURL = "http://localhost:5173/invitations"
	}

	bcryptCost := 12
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			bcryptCost = n
		}
	}

	return &Config{
		Env:                env,
		Port:               port,
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		CORSAllowedOrigins: corsAllowedOrigins,
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTokenTTL:     accessTTL,
		RefreshTokenTTL:    refreshTTL,
		InviteBaseURL:      inviteBaseURL,
		InviteTokenTTL:     inviteTTL,
		BcryptCost:         bcryptCost,
	}
}

func parseDuration(key string, defaultVal time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return defaultVal
	}
	return d
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
