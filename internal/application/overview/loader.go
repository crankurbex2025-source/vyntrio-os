package overview

import (
	"context"
	"fmt"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

const serviceStatusRunning = "running"

// ReadinessChecker evaluates process and dependency readiness.
type ReadinessChecker interface {
	Check(ctx context.Context) health.Result
}

// Loader assembles the authenticated overview response.
type Loader struct {
	repo        appsettings.Repository
	readiness   ReadinessChecker
	version     string
	commit      string
	environment string
	now         func() time.Time
}

// NewLoader creates an overview loader.
func NewLoader(
	repo appsettings.Repository,
	readiness ReadinessChecker,
	version, commit, environment string,
) Loader {
	return Loader{
		repo:        repo,
		readiness:   readiness,
		version:     version,
		commit:      commit,
		environment: environment,
		now:         time.Now,
	}
}

// Load returns the current safe overview DTO.
func (l Loader) Load(ctx context.Context) (Response, error) {
	host, err := l.repo.Get(ctx, setting.NamespaceSystem, setting.KeyHostname)
	if err != nil {
		return Response{}, fmt.Errorf("load system.hostname: %w", err)
	}
	displayName, err := setting.ValidateInstanceDisplayName(host.Value)
	if err != nil {
		return Response{}, fmt.Errorf("validate system.hostname: %w", err)
	}

	readiness := MapReadiness(l.readiness.Check(ctx))
	collectedAt := l.now().UTC().Format(time.RFC3339Nano)

	return Response{
		Instance: InstanceSection{
			Name:    displayName,
			Version: l.version,
			Commit:  l.commit,
		},
		API: APISection{
			Environment: l.environment,
		},
		Service: ServiceSection{
			Status: serviceStatusRunning,
		},
		Readiness:   readiness,
		CollectedAt: collectedAt,
	}, nil
}

// MapReadiness converts health readiness results into overview readiness fields.
func MapReadiness(result health.Result) ReadinessSection {
	if result.DatabaseOK {
		return ReadinessSection{
			Status:   "ready",
			Database: "ok",
		}
	}
	return ReadinessSection{
		Status:   "not_ready",
		Database: "error",
	}
}
