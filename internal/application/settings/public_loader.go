package settings

import (
	"context"
	"fmt"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

// PublicSettingsLoader assembles the safe owner-visible settings response from persistence.
type PublicSettingsLoader struct {
	repo        Repository
	version     string
	environment string
}

// NewPublicSettingsLoader creates a loader for GET /api/v1/settings.
func NewPublicSettingsLoader(repo Repository, version, environment string) PublicSettingsLoader {
	return PublicSettingsLoader{
		repo:        repo,
		version:     version,
		environment: environment,
	}
}

// Load returns the current safe public settings DTO.
func (l PublicSettingsLoader) Load(ctx context.Context) (PublicSettingsResponse, error) {
	host, err := l.repo.Get(ctx, setting.NamespaceSystem, setting.KeyHostname)
	if err != nil {
		return PublicSettingsResponse{}, fmt.Errorf("load system.hostname: %w", err)
	}
	displayName, err := setting.ValidateInstanceDisplayName(host.Value)
	if err != nil {
		return PublicSettingsResponse{}, fmt.Errorf("validate system.hostname: %w", err)
	}

	return PublicSettingsResponse{
		Instance: PublicInstanceSettings{
			Name:    displayName,
			Version: l.version,
		},
		API: PublicAPISettings{
			Environment: l.environment,
		},
	}, nil
}
