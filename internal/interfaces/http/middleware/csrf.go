package middleware

import (
	"net/http"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/auth"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

const csrfHeaderName = "X-CSRF-Token"

// RequireCSRF rejects authenticated requests without a valid session-bound CSRF header.
func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		csrf, ok := auth.SessionCSRFFromContext(r.Context())
		if !ok {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "CSRF validation failed", requestID)
			return
		}

		rawHeader := strings.TrimSpace(r.Header.Get(csrfHeaderName))
		if !csrf.Valid(rawHeader) {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "CSRF validation failed", requestID)
			return
		}

		next.ServeHTTP(w, r)
	})
}
