package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// SessionRepository implements appidentity.SessionStore.
type SessionRepository struct {
	q *sqlcgen.Queries
}

var _ appidentity.SessionStore = (*SessionRepository)(nil)

// NewSessionRepository creates a session store backed by the given database.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{q: sqlcgen.New(db)}
}

// CreateSession inserts a new session record.
func (r *SessionRepository) CreateSession(ctx context.Context, input appidentity.CreateSessionInput) error {
	if err := r.q.CreateSession(ctx, sqlcgen.CreateSessionParams{
		ID:               input.ID,
		SessionTokenHash: input.SessionTokenHash,
		UserID:           string(input.UserID),
		CsrfTokenHash:    input.CSRFTokenHash,
		ExpiresAt:        input.ExpiresAt,
		IdleExpiresAt:    input.IdleExpiresAt,
		CreatedAt:        input.CreatedAt,
		LastSeenAt:       input.LastSeenAt,
		UserAgentHash:    nullString(input.UserAgentHash),
		IpHash:           nullString(input.IPHash),
	}); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSessionByTokenHash returns internal session validation data including token hashes.
func (r *SessionRepository) GetSessionByTokenHash(ctx context.Context, sessionTokenHash string) (appidentity.SessionCredential, error) {
	row, err := r.q.GetSessionByTokenHash(ctx, sessionTokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appidentity.SessionCredential{}, appidentity.ErrNotFound
		}
		return appidentity.SessionCredential{}, fmt.Errorf("get session by token hash: %w", err)
	}
	return mapSessionCredential(row), nil
}

// TouchSession updates the last seen timestamp.
func (r *SessionRepository) TouchSession(ctx context.Context, id string, lastSeenAt string) error {
	if err := r.q.TouchSession(ctx, sqlcgen.TouchSessionParams{
		LastSeenAt: lastSeenAt,
		ID:         id,
	}); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

// RevokeSessionByID marks a session revoked.
func (r *SessionRepository) RevokeSessionByID(ctx context.Context, id string, revokedAt string) error {
	if err := r.q.RevokeSessionByID(ctx, sqlcgen.RevokeSessionByIDParams{
		RevokedAt: nullString(revokedAt),
		ID:        id,
	}); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeAllSessionsForUser marks all active sessions for a user revoked.
func (r *SessionRepository) RevokeAllSessionsForUser(ctx context.Context, userID domainidentity.UserID, revokedAt string) error {
	if err := r.q.RevokeAllSessionsForUser(ctx, sqlcgen.RevokeAllSessionsForUserParams{
		RevokedAt: nullString(revokedAt),
		UserID:    string(userID),
	}); err != nil {
		return fmt.Errorf("revoke all sessions for user: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes expired sessions only.
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	count, err := r.q.DeleteExpiredSessions(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return count, nil
}

func mapSessionCredential(row sqlcgen.GetSessionByTokenHashRow) appidentity.SessionCredential {
	return appidentity.SessionCredential{
		Session: appidentity.Session{
			ID:            row.ID,
			UserID:        domainidentity.UserID(row.UserID),
			CreatedAt:     row.CreatedAt,
			ExpiresAt:     row.ExpiresAt,
			IdleExpiresAt: row.IdleExpiresAt,
			LastSeenAt:    row.LastSeenAt,
			RevokedAt:     stringFromNull(row.RevokedAt),
		},
		SessionTokenHash: row.SessionTokenHash,
		CSRFTokenHash:    row.CsrfTokenHash,
		UserAgentHash:    stringFromNull(row.UserAgentHash),
		IPHash:           stringFromNull(row.IpHash),
	}
}
