package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

type verifyRepository struct {
	timezone setting.Setting
	hostname setting.Setting
}

func (m *verifyRepository) Get(_ context.Context, namespace, key string) (setting.Setting, error) {
	switch key {
	case setting.KeyTimezone:
		if m.timezone.Namespace != namespace {
			return setting.Setting{}, settings.ErrNotFound
		}
		return m.timezone, nil
	case setting.KeyHostname:
		if m.hostname.Namespace != namespace {
			return setting.Setting{}, settings.ErrNotFound
		}
		return m.hostname, nil
	default:
		return setting.Setting{}, settings.ErrNotFound
	}
}

func (m *verifyRepository) Set(context.Context, setting.Setting) error {
	return errors.New("not implemented")
}

func (m *verifyRepository) ListByNamespace(context.Context, string) ([]setting.Setting, error) {
	return nil, errors.New("not implemented")
}

func TestVerifyPersistedSuccess(t *testing.T) {
	repo := &verifyRepository{
		timezone: setting.Setting{
			Namespace: setting.NamespaceSystem,
			Key:       setting.KeyTimezone,
			Value:     "UTC",
		},
		hostname: setting.Setting{
			Namespace: setting.NamespaceSystem,
			Key:       setting.KeyHostname,
			Value:     "vyntrio",
		},
	}
	snap := settings.NewSnapshot(settings.SystemSettings{
		Timezone: "UTC",
		Hostname: "vyntrio",
	})

	if err := settings.VerifyPersisted(context.Background(), repo, snap); err != nil {
		t.Fatalf("VerifyPersisted() error: %v", err)
	}
}

func TestVerifyPersistedTimezoneMismatch(t *testing.T) {
	repo := &verifyRepository{
		timezone: setting.Setting{
			Namespace: setting.NamespaceSystem,
			Key:       setting.KeyTimezone,
			Value:     "Europe/Berlin",
		},
		hostname: setting.Setting{
			Namespace: setting.NamespaceSystem,
			Key:       setting.KeyHostname,
			Value:     "vyntrio",
		},
	}
	snap := settings.NewSnapshot(settings.SystemSettings{
		Timezone: "UTC",
		Hostname: "vyntrio",
	})

	err := settings.VerifyPersisted(context.Background(), repo, snap)
	if !errors.Is(err, settings.ErrPersistedMismatch) {
		t.Fatalf("expected ErrPersistedMismatch, got %v", err)
	}
}

func TestVerifyPersistedHostnameMismatch(t *testing.T) {
	repo := &verifyRepository{
		timezone: setting.Setting{
			Namespace: setting.NamespaceSystem,
			Key:       setting.KeyTimezone,
			Value:     "UTC",
		},
		hostname: setting.Setting{
			Namespace: setting.NamespaceSystem,
			Key:       setting.KeyHostname,
			Value:     "other-host",
		},
	}
	snap := settings.NewSnapshot(settings.SystemSettings{
		Timezone: "UTC",
		Hostname: "vyntrio",
	})

	err := settings.VerifyPersisted(context.Background(), repo, snap)
	if !errors.Is(err, settings.ErrPersistedMismatch) {
		t.Fatalf("expected ErrPersistedMismatch, got %v", err)
	}
}
