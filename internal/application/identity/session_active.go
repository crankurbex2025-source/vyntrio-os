package identity

import "time"

// IsSessionActive reports whether a session is eligible for authenticated use.
func IsSessionActive(session Session, now time.Time) bool {
	if session.RevokedAt != "" {
		return false
	}
	expiresAt, err := time.Parse(time.RFC3339, session.ExpiresAt)
	if err != nil {
		return false
	}
	idleExpiresAt, err := time.Parse(time.RFC3339, session.IdleExpiresAt)
	if err != nil {
		return false
	}
	now = now.UTC()
	return !now.After(expiresAt) && !now.After(idleExpiresAt)
}
