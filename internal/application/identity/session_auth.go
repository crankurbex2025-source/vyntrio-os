package identity

import (
	"context"
	"errors"
	"time"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

// MaxSessionCookieValueLen bounds raw session cookie values before hashing.
// 32-byte tokens encode to 43 characters; 64-byte tokens encode to 86 characters.
const MaxSessionCookieValueLen = 128

// MaxCSRFHeaderValueLen bounds raw CSRF header values before hashing.
const MaxCSRFHeaderValueLen = 128

// ResolvedSession holds authenticated subject fields resolved from a session cookie.
type ResolvedSession struct {
	UserID        domainidentity.UserID
	Role          domainidentity.Role
	CSRFTokenHash string
}

// SessionAuthRecord holds session and user fields required for authentication.
type SessionAuthRecord struct {
	SessionID     string
	UserID        domainidentity.UserID
	CSRFTokenHash string
	ExpiresAt     string
	IdleExpiresAt string
	RevokedAt     string
	UserStatus    UserStatus
	Role          domainidentity.Role
}

// SessionAuthStore looks up authentication state by session token hash.
type SessionAuthStore interface {
	GetSessionAuthByTokenHash(ctx context.Context, sessionTokenHash string) (SessionAuthRecord, error)
}

// SessionResolver resolves a raw session cookie to an authenticated subject.
type SessionResolver struct {
	store SessionAuthStore
	now   func() time.Time
}

// NewSessionResolver returns a session resolver.
func NewSessionResolver(store SessionAuthStore) *SessionResolver {
	return &SessionResolver{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// Resolve validates a raw session cookie value and returns an authenticated subject.
// Missing or invalid credentials yield ok=false with nil error.
// Store failures other than not-found return an error.
func (r *SessionResolver) Resolve(
	ctx context.Context,
	rawSessionToken string,
) (ResolvedSession, bool, error) {
	if err := ctx.Err(); err != nil {
		return ResolvedSession{}, false, err
	}
	if rawSessionToken == "" || len(rawSessionToken) > MaxSessionCookieValueLen {
		return ResolvedSession{}, false, nil
	}

	record, err := r.store.GetSessionAuthByTokenHash(ctx, HashRawToken(rawSessionToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ResolvedSession{}, false, nil
		}
		return ResolvedSession{}, false, err
	}
	if !isSessionAuthValid(record, r.now()) {
		return ResolvedSession{}, false, nil
	}
	return ResolvedSession{
		UserID:        record.UserID,
		Role:          record.Role,
		CSRFTokenHash: record.CSRFTokenHash,
	}, true, nil
}

func isSessionAuthValid(record SessionAuthRecord, now time.Time) bool {
	if record.RevokedAt != "" {
		return false
	}
	if record.CSRFTokenHash == "" {
		return false
	}
	if record.UserStatus != UserStatusActive {
		return false
	}
	if !record.Role.Valid() {
		return false
	}
	expiresAt, err := time.Parse(time.RFC3339, record.ExpiresAt)
	if err != nil {
		return false
	}
	idleExpiresAt, err := time.Parse(time.RFC3339, record.IdleExpiresAt)
	if err != nil {
		return false
	}
	now = now.UTC()
	return now.Before(expiresAt) && now.Before(idleExpiresAt)
}
