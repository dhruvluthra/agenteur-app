package middleware

import (
	"net/http"
	"strings"
)

// RequireJSONContentType rejects POST/PUT/PATCH/DELETE requests that don't
// carry Content-Type: application/json. This serves as CSRF mitigation —
// HTML forms cannot send JSON, so cross-origin form submissions are blocked.
func RequireJSONContentType() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				ct := r.Header.Get("Content-Type")
				if !strings.HasPrefix(ct, "application/json") {
					http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
