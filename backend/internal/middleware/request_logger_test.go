package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLoggerLogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health?x=1", nil)
	req.RemoteAddr = "127.0.0.1:5000"
	req.Header.Set("User-Agent", "go-test")
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "rid-1"))
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected a log line")
	}

	entry := map[string]any{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse log line: %v", err)
	}

	if msg, _ := entry["msg"].(string); msg != "http request" {
		t.Fatalf("expected msg 'http request', got %q", msg)
	}
	if method, _ := entry["method"].(string); method != http.MethodGet {
		t.Fatalf("expected method GET, got %q", method)
	}
	if path, _ := entry["path"].(string); path != "/health" {
		t.Fatalf("expected path /health, got %q", path)
	}
	if query, _ := entry["query"].(string); query != "x=1" {
		t.Fatalf("expected query x=1, got %q", query)
	}
	if status, _ := entry["status"].(float64); int(status) != http.StatusCreated {
		t.Fatalf("expected status 201, got %v", entry["status"])
	}
	if rid, _ := entry["request_id"].(string); rid != "rid-1" {
		t.Fatalf("expected request_id rid-1, got %q", rid)
	}
}

func TestRequestLoggerUsesErrorLevelFor5xx(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "rid-2"))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	line := strings.TrimSpace(buf.String())
	entry := map[string]any{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse log line: %v", err)
	}

	if level, _ := entry["level"].(string); level != "ERROR" {
		t.Fatalf("expected ERROR level, got %q", level)
	}
}
