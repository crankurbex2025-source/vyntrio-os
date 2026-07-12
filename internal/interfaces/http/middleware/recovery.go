package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// Recovery catches panics and returns a JSON 500 error.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := GetRequestID(r.Context())
					logger.Error("panic recovered",
						"error", rec,
						"request_id", requestID,
						"stack", string(debug.Stack()),
					)
					response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR",
						"An internal error occurred", requestID)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
