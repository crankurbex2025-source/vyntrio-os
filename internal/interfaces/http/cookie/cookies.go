// Package cookie provides session cookie helpers for browser authentication.
package cookie

import (
	"net/http"
	"time"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
)

const (
	// SessionCookieName is the HttpOnly session cookie (ADR-0004).
	SessionCookieName = "vyntrio_session"
	sessionCookiePath = "/"
)

// Policy controls secure cookie transport attributes.
type Policy struct {
	Secure bool
}

// NewPolicy returns cookie transport settings from runtime configuration.
func NewPolicy(cookieSecure bool) Policy {
	return Policy{Secure: cookieSecure}
}

// SetSessionCookie writes the session cookie for a successful login.
func (p Policy) SetSessionCookie(w http.ResponseWriter, material appidentity.SessionMaterial, now time.Time) {
	maxAge := cookieMaxAge(material.ExpiresAt, now)
	expires := material.ExpiresAt.UTC()

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    material.RawSessionToken,
		Path:     sessionCookiePath,
		MaxAge:   maxAge,
		Expires:  expires,
		HttpOnly: true,
		Secure:   p.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearSessionCookie removes the session cookie from the client.
func (p Policy) ClearSessionCookie(w http.ResponseWriter) {
	expired := time.Unix(0, 0).UTC()
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     sessionCookiePath,
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: true,
		Secure:   p.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func cookieMaxAge(expiresAt, now time.Time) int {
	seconds := int(expiresAt.Sub(now.UTC()).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
