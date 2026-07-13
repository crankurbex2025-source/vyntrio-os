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

// SessionAuthRepository implements appidentity.SessionAuthStore.
type SessionAuthRepository struct {
	q *sqlcgen.Queries
}

var _ appidentity.SessionAuthStore = (*SessionAuthRepository)(nil)

// NewSessionAuthRepository creates a session authentication lookup store.
func NewSessionAuthRepository(db *sql.DB) *SessionAuthRepository {
	return &SessionAuthRepository{q: sqlcgen.New(db)}
}

// GetSessionAuthByTokenHash returns session and user fields required for authentication.
func (r *SessionAuthRepository) GetSessionAuthByTokenHash(
	ctx context.Context,
	sessionTokenHash string,
) (appidentity.SessionAuthRecord, error) {
	row, err := r.q.GetSessionAuthByTokenHash(ctx, sessionTokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appidentity.SessionAuthRecord{}, appidentity.ErrNotFound
		}
		return appidentity.SessionAuthRecord{}, fmt.Errorf("get session auth by token hash: %w", err)
	}
	return mapSessionAuthRecord(row), nil
}

func mapSessionAuthRecord(row sqlcgen.GetSessionAuthByTokenHashRow) appidentity.SessionAuthRecord {
	return appidentity.SessionAuthRecord{
		SessionID:     row.SessionID,
		UserID:        domainidentity.UserID(row.UserID),
		ExpiresAt:     row.ExpiresAt,
		IdleExpiresAt: row.IdleExpiresAt,
		RevokedAt:     stringFromNull(row.RevokedAt),
		UserStatus:    appidentity.UserStatus(row.UserStatus),
		Role:          domainidentity.Role(row.UserRole),
	}
}
