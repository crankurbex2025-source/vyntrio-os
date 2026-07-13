package identity

import (
	"context"
	"time"
)

// LogoutSessionRevoker revokes an active session and appends a logout audit event atomically.
type LogoutSessionRevoker interface {
	RevokeActiveSessionByTokenHash(
		ctx context.Context,
		sessionTokenHash, revokedAt string,
		audit AppendSecurityAuditEventInput,
	) (revoked bool, err error)
}

// LogoutService revokes the current session when present.
type LogoutService struct {
	revoker LogoutSessionRevoker
	now     func() time.Time
}

// NewLogoutService returns a logout service.
func NewLogoutService(revoker LogoutSessionRevoker) *LogoutService {
	return &LogoutService{
		revoker: revoker,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

// LogoutResult describes server-side logout effects.
type LogoutResult struct {
	SessionRevoked bool
	RevokeFailed   bool
}

// Logout revokes the session identified by the raw session cookie value when active.
func (s *LogoutService) Logout(ctx context.Context, rawSessionToken, auditID string) (LogoutResult, error) {
	if err := ctx.Err(); err != nil {
		return LogoutResult{}, err
	}
	if rawSessionToken == "" {
		return LogoutResult{}, nil
	}

	tokenHash := HashRawToken(rawSessionToken)
	revoked, err := s.revoker.RevokeActiveSessionByTokenHash(ctx, tokenHash, FormatUTCTime(s.now()), AppendSecurityAuditEventInput{
		ID:           auditID,
		EventType:    "identity.logout.succeeded",
		Result:       "success",
		MetadataJSON: `{}`,
	})
	if err != nil {
		return LogoutResult{RevokeFailed: true}, err
	}
	return LogoutResult{SessionRevoked: revoked}, nil
}
