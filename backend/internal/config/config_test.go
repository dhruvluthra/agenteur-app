package config

import (
	"reflect"
	"testing"
)

func TestParseCSVEnv(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.com, http://b.com, ,http://c.com,")

	got := parseCSVEnv("CORS_ALLOWED_ORIGINS")
	want := []string{"http://a.com", "http://b.com", "http://c.com"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCSVEnv mismatch: got %v, want %v", got, want)
	}
}

func TestParseCSVEnvEmpty(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "   ")

	got := parseCSVEnv("CORS_ALLOWED_ORIGINS")
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestLoadParsesCORSAllowedOrigins(t *testing.T) {
	t.Setenv("ENV", "dev")
	t.Setenv("PORT", ":9090")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")

	cfg := Load()

	want := []string{"http://localhost:5173", "http://localhost:3000"}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, want) {
		t.Fatalf("CORSAllowedOrigins mismatch: got %v, want %v", cfg.CORSAllowedOrigins, want)
	}
}
