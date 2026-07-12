package sqlite_test

import (
	"context"
	"errors"
	"testing"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestSettingsRepositorySeededDefaults(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	repo := sqlite.NewSettingsRepository(store.DB())

	tz, err := repo.Get(context.Background(), setting.NamespaceSystem, setting.KeyTimezone)
	if err != nil {
		t.Fatalf("Get(timezone) error: %v", err)
	}
	if tz.Value != "UTC" {
		t.Fatalf("timezone = %q, want UTC", tz.Value)
	}

	host, err := repo.Get(context.Background(), setting.NamespaceSystem, setting.KeyHostname)
	if err != nil {
		t.Fatalf("Get(hostname) error: %v", err)
	}
	if host.Value != "vyntrio" {
		t.Fatalf("hostname = %q, want vyntrio", host.Value)
	}
}

func TestSettingsRepositoryUpsertAndList(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	repo := sqlite.NewSettingsRepository(store.DB())
	ctx := context.Background()

	if err := repo.Set(ctx, setting.Setting{
		Namespace: setting.NamespaceSystem,
		Key:       setting.KeyHostname,
		Value:     "vyntrio-test",
		ValueType: setting.ValueTypeString,
	}); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := repo.Get(ctx, setting.NamespaceSystem, setting.KeyHostname)
	if err != nil {
		t.Fatalf("Get() after Set error: %v", err)
	}
	if got.Value != "vyntrio-test" {
		t.Fatalf("hostname = %q, want vyntrio-test", got.Value)
	}

	all, err := repo.ListByNamespace(ctx, setting.NamespaceSystem)
	if err != nil {
		t.Fatalf("ListByNamespace() error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("len(settings) = %d, want 2", len(all))
	}
}

func TestSettingsRepositoryRejectsInvalidKey(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	repo := sqlite.NewSettingsRepository(store.DB())
	err = repo.Set(context.Background(), setting.Setting{
		Namespace: setting.NamespaceSystem,
		Key:       "locale",
		Value:     "de-DE",
		ValueType: setting.ValueTypeString,
	})
	if !errors.Is(err, setting.ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestSettingsRepositoryGetNotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if _, err := store.DB().ExecContext(ctx,
		`DELETE FROM settings WHERE namespace = ? AND key = ?`,
		setting.NamespaceSystem, setting.KeyTimezone,
	); err != nil {
		t.Fatalf("delete seed timezone: %v", err)
	}

	repo := sqlite.NewSettingsRepository(store.DB())
	_, err = repo.Get(ctx, setting.NamespaceSystem, setting.KeyTimezone)
	if !errors.Is(err, appsettings.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSettingsRepositoryGetRejectsInvalidKey(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	repo := sqlite.NewSettingsRepository(store.DB())
	_, err = repo.Get(context.Background(), setting.NamespaceSystem, "locale")
	if !errors.Is(err, setting.ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestSettingsRepositoryRejectsInvalidNamespace(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	repo := sqlite.NewSettingsRepository(store.DB())
	_, err = repo.ListByNamespace(context.Background(), "network")
	if !errors.Is(err, setting.ErrInvalidNamespace) {
		t.Fatalf("expected ErrInvalidNamespace, got %v", err)
	}
}
