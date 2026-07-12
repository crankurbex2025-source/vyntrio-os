package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// SettingsRepository implements settings.Repository using sqlc-generated queries.
type SettingsRepository struct {
	q *sqlcgen.Queries
}

// NewSettingsRepository creates a settings repository backed by the given database.
func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{q: sqlcgen.New(db)}
}

// Get returns a setting by namespace and key.
func (r *SettingsRepository) Get(ctx context.Context, namespace, key string) (setting.Setting, error) {
	if err := setting.ValidateKey(namespace, key); err != nil {
		return setting.Setting{}, err
	}

	row, err := r.q.GetSetting(ctx, sqlcgen.GetSettingParams{
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return setting.Setting{}, appsettings.ErrNotFound
		}
		return setting.Setting{}, fmt.Errorf("get setting: %w", err)
	}

	return toDomainSetting(row), nil
}

// Set validates and upserts a setting.
func (r *SettingsRepository) Set(ctx context.Context, s setting.Setting) error {
	if err := s.Validate(); err != nil {
		return err
	}

	if err := r.q.UpsertSetting(ctx, sqlcgen.UpsertSettingParams{
		Namespace: s.Namespace,
		Key:       s.Key,
		Value:     s.Value,
		ValueType: string(s.ValueType),
	}); err != nil {
		return fmt.Errorf("upsert setting: %w", err)
	}
	return nil
}

// ListByNamespace returns all settings in a namespace.
func (r *SettingsRepository) ListByNamespace(ctx context.Context, namespace string) ([]setting.Setting, error) {
	if err := setting.ValidateNamespace(namespace); err != nil {
		return nil, err
	}

	rows, err := r.q.ListSettingsByNamespace(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}

	out := make([]setting.Setting, 0, len(rows))
	for _, row := range rows {
		if err := setting.ValidateKey(row.Namespace, row.Key); err != nil {
			return nil, fmt.Errorf("list settings: %w", err)
		}
		out = append(out, toDomainSetting(row))
	}
	return out, nil
}

func toDomainSetting(row sqlcgen.Setting) setting.Setting {
	return setting.Setting{
		Namespace: row.Namespace,
		Key:       row.Key,
		Value:     row.Value,
		ValueType: setting.ValueType(row.ValueType),
		UpdatedAt: row.UpdatedAt,
	}
}
