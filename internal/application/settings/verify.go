package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

// ErrPersistedMismatch indicates persisted settings differ from the startup snapshot.
var ErrPersistedMismatch = errors.New("persisted settings mismatch")

// VerifyPersisted re-reads system settings once and compares exact values to snap.
func VerifyPersisted(ctx context.Context, repo Repository, snap Snapshot) error {
	tz, err := repo.Get(ctx, setting.NamespaceSystem, setting.KeyTimezone)
	if err != nil {
		return fmt.Errorf("verify system.timezone: %w", err)
	}
	if tz.Value != snap.Timezone {
		return fmt.Errorf("%w: system.timezone expected %q got %q", ErrPersistedMismatch, snap.Timezone, tz.Value)
	}

	host, err := repo.Get(ctx, setting.NamespaceSystem, setting.KeyHostname)
	if err != nil {
		return fmt.Errorf("verify system.hostname: %w", err)
	}
	if host.Value != snap.Hostname {
		return fmt.Errorf("%w: system.hostname expected %q got %q", ErrPersistedMismatch, snap.Hostname, host.Value)
	}

	return nil
}
