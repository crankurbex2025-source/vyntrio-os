// Package response provides JSON HTTP response helpers.
package response

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the standard JSON error envelope (docs/API_CONVENTIONS.md).
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes a client-visible error.
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// Error writes a standard error JSON response.
func Error(w http.ResponseWriter, status int, code, message, requestID string) {
	JSON(w, status, ErrorBody{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}

// NoContent writes an empty success response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
