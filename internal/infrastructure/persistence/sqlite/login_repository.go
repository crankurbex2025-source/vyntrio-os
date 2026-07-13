package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// LoginRepository atomically creates sessions and login audit events.
type LoginRepository struct {
	db *sql.DB
}

// NewLoginRepository creates a login repository.
func NewLoginRepository(db *sql.DB) *LoginRepository {
	return &LoginRepository{db: db}
}

var _ appidentity.LoginSessionCreator = (*LoginRepository)(nil)

// CreateSessionWithAudit inserts a session and login audit event in one transaction.
func (r *LoginRepository) CreateSessionWithAudit(
	ctx context.Context,
	session appidentity.CreateSessionInput,
	audit appidentity.AppendSecurityAuditEventInput,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin login transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := sqlcgen.New(tx)

	if err := q.CreateSession(ctx, sqlcgen.CreateSessionParams{
		ID:               session.ID,
		SessionTokenHash: session.SessionTokenHash,
		UserID:           string(session.UserID),
		CsrfTokenHash:    session.CSRFTokenHash,
		ExpiresAt:        session.ExpiresAt,
		IdleExpiresAt:    session.IdleExpiresAt,
		CreatedAt:        session.CreatedAt,
		LastSeenAt:       session.LastSeenAt,
		UserAgentHash:    nullString(session.UserAgentHash),
		IpHash:           nullString(session.IPHash),
	}); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	if err := q.AppendSecurityAuditEvent(ctx, sqlcgen.AppendSecurityAuditEventParams{
		ID:            audit.ID,
		ActorUserID:   nullString(string(audit.ActorUserID)),
		SubjectUserID: nullString(string(audit.SubjectUserID)),
		EventType:     audit.EventType,
		Result:        audit.Result,
		MetadataJson:  nullString(audit.MetadataJSON),
	}); err != nil {
		return fmt.Errorf("append login audit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login transaction: %w", err)
	}
	return nil
}
