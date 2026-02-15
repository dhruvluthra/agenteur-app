package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDGeneratesAndSetsHeader(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := GetRequestID(r.Context()); id == "" {
			t.Fatal("expected request id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected response request id header")
	}
}

func TestRequestIDPreservesIncomingHeader(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := GetRequestID(r.Context()); got != "abc-123" {
			t.Fatalf("expected context request id abc-123, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "abc-123")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get(RequestIDHeader); got != "abc-123" {
		t.Fatalf("expected response request id abc-123, got %q", got)
	}
}
