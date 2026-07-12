package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

type mockRepository struct {
	byKey map[string]setting.Setting
}

func (m *mockRepository) Get(_ context.Context, namespace, key string) (setting.Setting, error) {
	s, ok := m.byKey[key]
	if !ok {
		return setting.Setting{}, settings.ErrNotFound
	}
	if s.Namespace != namespace {
		return setting.Setting{}, settings.ErrNotFound
	}
	return s, nil
}

func (m *mockRepository) Set(context.Context, setting.Setting) error {
	return errors.New("not implemented")
}

func (m *mockRepository) ListByNamespace(context.Context, string) ([]setting.Setting, error) {
	return nil, errors.New("not implemented")
}

func TestLoadSystemSettingsSuccess(t *testing.T) {
	repo := &mockRepository{
		byKey: map[string]setting.Setting{
			setting.KeyTimezone: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyTimezone,
				Value:     "UTC",
				ValueType: setting.ValueTypeString,
			},
			setting.KeyHostname: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyHostname,
				Value:     "vyntrio",
				ValueType: setting.ValueTypeString,
			},
		},
	}

	got, err := settings.NewReader(repo).LoadSystemSettings(context.Background())
	if err != nil {
		t.Fatalf("LoadSystemSettings() error: %v", err)
	}
	if got.Timezone != "UTC" || got.Hostname != "vyntrio" {
		t.Fatalf("LoadSystemSettings() = %+v, want timezone=UTC hostname=vyntrio", got)
	}
}

func TestLoadSystemSettingsMissingTimezone(t *testing.T) {
	repo := &mockRepository{
		byKey: map[string]setting.Setting{
			setting.KeyHostname: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyHostname,
				Value:     "vyntrio",
				ValueType: setting.ValueTypeString,
			},
		},
	}

	_, err := settings.NewReader(repo).LoadSystemSettings(context.Background())
	if !errors.Is(err, settings.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLoadSystemSettingsInvalidHostname(t *testing.T) {
	repo := &mockRepository{
		byKey: map[string]setting.Setting{
			setting.KeyTimezone: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyTimezone,
				Value:     "UTC",
				ValueType: setting.ValueTypeString,
			},
			setting.KeyHostname: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyHostname,
				Value:     "bad host",
				ValueType: setting.ValueTypeString,
			},
		},
	}

	_, err := settings.NewReader(repo).LoadSystemSettings(context.Background())
	if !errors.Is(err, setting.ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}
