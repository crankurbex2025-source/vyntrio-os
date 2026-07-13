package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// InstanceDisplayNameRepository atomically updates instance display name settings with audit events.
type InstanceDisplayNameRepository struct {
	db *sql.DB
}

// NewInstanceDisplayNameRepository creates an instance display name repository.
func NewInstanceDisplayNameRepository(db *sql.DB) *InstanceDisplayNameRepository {
	return &InstanceDisplayNameRepository{db: db}
}

var _ appsettings.InstanceDisplayNameStore = (*InstanceDisplayNameRepository)(nil)

// UpdateInstanceDisplayNameWithAudit updates system.hostname when changed and appends a success audit event.
func (r *InstanceDisplayNameRepository) UpdateInstanceDisplayNameWithAudit(
	ctx context.Context,
	displayName string,
	actorUserID domainidentity.UserID,
	auditID string,
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin instance display name transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := sqlcgen.New(tx)

	current, err := q.GetSetting(ctx, sqlcgen.GetSettingParams{
		Namespace: setting.NamespaceSystem,
		Key:       setting.KeyHostname,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, appsettings.ErrNotFound
		}
		return false, fmt.Errorf("get system.hostname: %w", err)
	}

	if current.Value == displayName {
		return false, nil
	}

	if err := q.UpsertSetting(ctx, sqlcgen.UpsertSettingParams{
		Namespace: setting.NamespaceSystem,
		Key:       setting.KeyHostname,
		Value:     displayName,
		ValueType: string(setting.ValueTypeString),
	}); err != nil {
		return false, fmt.Errorf("upsert system.hostname: %w", err)
	}

	if err := q.AppendSecurityAuditEvent(ctx, sqlcgen.AppendSecurityAuditEventParams{
		ID:            auditID,
		ActorUserID:   nullString(string(actorUserID)),
		SubjectUserID: nullString(string(actorUserID)),
		EventType:     appsettings.AuditEventInstanceDisplayNameUpdated,
		Result:        "success",
		MetadataJson:  nullString(`{}`),
	}); err != nil {
		return false, fmt.Errorf("append instance display name audit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit instance display name transaction: %w", err)
	}
	return true, nil
}
