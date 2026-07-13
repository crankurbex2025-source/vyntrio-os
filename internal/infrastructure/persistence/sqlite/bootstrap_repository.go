package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// BootstrapRepository atomically creates the first owner and audit event.
type BootstrapRepository struct {
	db *sql.DB
}

// NewBootstrapRepository creates a bootstrap repository.
func NewBootstrapRepository(db *sql.DB) *BootstrapRepository {
	return &BootstrapRepository{db: db}
}

var _ appidentity.BootstrapCreator = (*BootstrapRepository)(nil)

// CreateFirstOwner inserts the first user and audit event in one transaction.
func (r *BootstrapRepository) CreateFirstOwner(
	ctx context.Context,
	user appidentity.BootstrapCreateInput,
	audit appidentity.BootstrapAuditInput,
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin bootstrap transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := sqlcgen.New(tx)

	rows, err := q.CreateUserIfNoUsersExist(ctx, sqlcgen.CreateUserIfNoUsersExistParams{
		ID:                 user.UserID,
		Username:           user.Username,
		DisplayName:        sql.NullString{},
		PasswordHash:       user.PasswordHash,
		Role:               user.Role,
		Status:             user.Status,
		MustChangePassword: intFromBool(user.MustChangePassword),
	})
	if err != nil {
		return false, fmt.Errorf("create first user: %w", err)
	}
	if rows == 0 {
		return false, nil
	}

	if err := q.AppendSecurityAuditEvent(ctx, sqlcgen.AppendSecurityAuditEventParams{
		ID:            audit.ID,
		ActorUserID:   nullString(audit.ActorUserID),
		SubjectUserID: nullString(audit.SubjectUserID),
		EventType:     audit.EventType,
		Result:        audit.Result,
		MetadataJson:  nullString(audit.MetadataJSON),
	}); err != nil {
		return false, fmt.Errorf("append bootstrap audit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit bootstrap transaction: %w", err)
	}
	return true, nil
}
