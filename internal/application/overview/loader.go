package overview

import (
	"context"
	"fmt"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/hostmetrics"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/netpresence"
)

const serviceStatusRunning = "running"

// ReadinessChecker evaluates process and dependency readiness.
type ReadinessChecker interface {
	Check(ctx context.Context) health.Result
}

// HostMetricsCollector assembles read-only host metrics for the overview.
type HostMetricsCollector interface {
	Collect(ctx context.Context) hostmetrics.Host
}

// BackupStatusLoader reads the sanitized backup status read model.
type BackupStatusLoader interface {
	Read(ctx context.Context) backupstatus.Backup
}

// NetworkPresenceCollector assembles read-only network presence for the overview.
type NetworkPresenceCollector interface {
	Collect(ctx context.Context) netpresence.Network
}

// Loader assembles the authenticated overview response.
type Loader struct {
	repo            appsettings.Repository
	readiness       ReadinessChecker
	hostMetrics     HostMetricsCollector
	backupStatus    BackupStatusLoader
	networkPresence NetworkPresenceCollector
	version         string
	commit          string
	environment     string
	now             func() time.Time
}

// NewLoader creates an overview loader.
func NewLoader(
	repo appsettings.Repository,
	readiness ReadinessChecker,
	hostMetrics HostMetricsCollector,
	backupStatus BackupStatusLoader,
	networkPresence NetworkPresenceCollector,
	version, commit, environment string,
) Loader {
	return Loader{
		repo:            repo,
		readiness:       readiness,
		hostMetrics:     hostMetrics,
		backupStatus:    backupStatus,
		networkPresence: networkPresence,
		version:         version,
		commit:          commit,
		environment:     environment,
		now:             time.Now,
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
	service := ServiceSection{Status: serviceStatusRunning}
	backup := l.backupStatus.Read(ctx)
	runtime := AssembleRuntime(readiness, service)
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
		Service:     service,
		Readiness:   readiness,
		Host:        l.hostMetrics.Collect(ctx),
		Backup:      backup,
		Network:     l.networkPresence.Collect(ctx),
		Software:    AssembleSoftware(l.version, l.commit, l.environment),
		Runtime:     runtime,
		Health:      AssembleHealth(runtime, backup),
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
