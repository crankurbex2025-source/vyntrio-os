package settings

import (
	"context"
	"fmt"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

// SystemSettings holds validated system namespace settings loaded at startup.
type SystemSettings struct {
	Timezone string
	Hostname string
}

// Reader loads and validates internal system settings (read-only).
type Reader struct {
	repo Repository
}

// NewReader creates a settings reader backed by the given repository.
func NewReader(repo Repository) *Reader {
	return &Reader{repo: repo}
}

// LoadSystemSettings reads system.timezone and system.hostname once, validating each.
func (r *Reader) LoadSystemSettings(ctx context.Context) (SystemSettings, error) {
	tz, err := r.repo.Get(ctx, setting.NamespaceSystem, setting.KeyTimezone)
	if err != nil {
		return SystemSettings{}, fmt.Errorf("load system.timezone: %w", err)
	}
	if err := tz.Validate(); err != nil {
		return SystemSettings{}, fmt.Errorf("validate system.timezone: %w", err)
	}

	host, err := r.repo.Get(ctx, setting.NamespaceSystem, setting.KeyHostname)
	if err != nil {
		return SystemSettings{}, fmt.Errorf("load system.hostname: %w", err)
	}
	if err := host.Validate(); err != nil {
		return SystemSettings{}, fmt.Errorf("validate system.hostname: %w", err)
	}

	return SystemSettings{
		Timezone: tz.Value,
		Hostname: host.Value,
	}, nil
}
