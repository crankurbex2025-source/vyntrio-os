package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// LogoutRepository atomically revokes active sessions and appends logout audit events.
type LogoutRepository struct {
	db  *sql.DB
	now func() time.Time
}

// NewLogoutRepository creates a logout repository.
func NewLogoutRepository(db *sql.DB) *LogoutRepository {
	return &LogoutRepository{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
}

var _ appidentity.LogoutSessionRevoker = (*LogoutRepository)(nil)

// RevokeActiveSessionByTokenHash revokes one active session and appends an audit event.
func (r *LogoutRepository) RevokeActiveSessionByTokenHash(
	ctx context.Context,
	sessionTokenHash, revokedAt string,
	audit appidentity.AppendSecurityAuditEventInput,
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin logout transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := sqlcgen.New(tx)

	row, err := q.GetSessionByTokenHash(ctx, sessionTokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get session by token hash: %w", err)
	}

	session := appidentity.Session{
		ID:            row.ID,
		UserID:        domainidentity.UserID(row.UserID),
		ExpiresAt:     row.ExpiresAt,
		IdleExpiresAt: row.IdleExpiresAt,
		RevokedAt:     stringFromNull(row.RevokedAt),
	}
	if !appidentity.IsSessionActive(session, r.now()) {
		return false, nil
	}

	if err := q.RevokeSessionByID(ctx, sqlcgen.RevokeSessionByIDParams{
		RevokedAt: nullString(revokedAt),
		ID:        row.ID,
	}); err != nil {
		return false, fmt.Errorf("revoke session: %w", err)
	}

	userID := domainidentity.UserID(row.UserID)
	if err := q.AppendSecurityAuditEvent(ctx, sqlcgen.AppendSecurityAuditEventParams{
		ID:            audit.ID,
		ActorUserID:   nullString(string(userID)),
		SubjectUserID: nullString(string(userID)),
		EventType:     audit.EventType,
		Result:        audit.Result,
		MetadataJson:  nullString(audit.MetadataJSON),
	}); err != nil {
		return false, fmt.Errorf("append logout audit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit logout transaction: %w", err)
	}
	return true, nil
}
