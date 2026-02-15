package app

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"agenteur.ai/api/internal/config"
	imiddleware "agenteur.ai/api/internal/middleware"
)

func TestNewRouterHealthIncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := &config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}}

	h := NewRouter(cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if got := res.Header().Get(imiddleware.RequestIDHeader); got == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestNewRouterCORSPreflight(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := &config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}}

	h := NewRouter(cfg, logger)

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
}
