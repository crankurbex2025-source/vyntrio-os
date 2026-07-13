package auth

import (
	"context"
	"crypto/subtle"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
)

type sessionCSRFContextKey struct{}

// SessionCSRF holds server-derived CSRF verification material for the active session.
// The expected hash is not exported; callers validate raw header values via Valid only.
type SessionCSRF struct {
	expectedHashHex string
}

// NewSessionCSRF creates CSRF verification state from the authenticated session hash.
func NewSessionCSRF(expectedHashHex string) SessionCSRF {
	return SessionCSRF{expectedHashHex: expectedHashHex}
}

// Valid reports whether rawHeader matches the session-bound CSRF hash using constant-time comparison.
func (s SessionCSRF) Valid(rawHeader string) bool {
	if rawHeader == "" || len(rawHeader) > appidentity.MaxCSRFHeaderValueLen {
		return false
	}
	received := appidentity.HashRawToken(rawHeader)
	if len(received) != len(s.expectedHashHex) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(received), []byte(s.expectedHashHex)) == 1
}

// WithSessionCSRF stores session CSRF verification material in context.
func WithSessionCSRF(ctx context.Context, csrf SessionCSRF) context.Context {
	return context.WithValue(ctx, sessionCSRFContextKey{}, csrf)
}

// SessionCSRFFromContext returns session CSRF verification material when present.
func SessionCSRFFromContext(ctx context.Context) (SessionCSRF, bool) {
	value := ctx.Value(sessionCSRFContextKey{})
	if value == nil {
		return SessionCSRF{}, false
	}
	csrf, ok := value.(SessionCSRF)
	if !ok || csrf.expectedHashHex == "" {
		return SessionCSRF{}, false
	}
	return csrf, true
}
