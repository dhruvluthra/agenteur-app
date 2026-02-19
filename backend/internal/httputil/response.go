package httputil

import (
	"encoding/json"
	"net/http"
)

// JSON writes a success response with the standard envelope.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"data": data})
}

// ErrorResponse represents the error envelope.
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Error writes an error response with the standard envelope.
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError writes a 422 validation error response with field-level details.
func ValidationError(w http.ResponseWriter, message string, details map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}
