package backup

import (
	"context"
	"time"
)

// ServiceController abstracts systemd service lifecycle for tests.
type ServiceController interface {
	IsActive(ctx context.Context) (bool, error)
	Stop(ctx context.Context) error
	IsInactive(ctx context.Context) (bool, error)
	Start(ctx context.Context) error
}

// HealthProber checks local loopback health endpoints.
type HealthProber interface {
	Probe(ctx context.Context) error
}

// VersionFetcher reads API build metadata while the service is active.
type VersionFetcher interface {
	Fetch(ctx context.Context) (ReleaseMetadata, error)
}

// MigrationReader reads schema migration metadata from the database file.
type MigrationReader interface {
	Read(ctx context.Context, dbPath string) (ReleaseMetadata, error)
}

// RootChecker reports whether the effective UID is root.
type RootChecker interface {
	IsRoot() bool
}

// Clock provides time for deterministic tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// RealClock returns a Clock backed by time.Now.
func RealClock() Clock { return realClock{} }
