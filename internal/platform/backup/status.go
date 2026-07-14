package backup

import (
	"context"
	"errors"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

// RunStatusPublisher writes the backup status sidecar after completed runs.
type RunStatusPublisher interface {
	PublishSucceeded(ctx context.Context, stateRoot string, completedAt time.Time) error
	PublishFailed(ctx context.Context, stateRoot string, completedAt time.Time, failureClass string) error
}

// FailureClassFromCategory maps operator failure categories to status failure classes.
func FailureClassFromCategory(category FailureCategory) string {
	switch category {
	case FailureArtifact:
		return backupstatus.FailureArtifact
	case FailureRestart:
		return backupstatus.FailureRestart
	case FailureHealth:
		return backupstatus.FailureHealth
	case FailureReadiness:
		return backupstatus.FailureReadiness
	default:
		return backupstatus.FailureInternal
	}
}

// FailureClassFromError maps backup errors after quiesce to status failure classes.
func FailureClassFromError(err error) string {
	switch {
	case errors.Is(err, ErrSourceInvalid),
		errors.Is(err, ErrDestinationUnsafe),
		errors.Is(err, ErrArtifactFailed),
		errors.Is(err, ErrArtifactCollision):
		return backupstatus.FailureArtifact
	case errors.Is(err, ErrServiceStartFailed):
		return backupstatus.FailureRestart
	case errors.Is(err, ErrHealthProbeFailed):
		return backupstatus.FailureHealth
	case errors.Is(err, ErrReadinessProbeFailed):
		return backupstatus.FailureReadiness
	default:
		return backupstatus.FailureInternal
	}
}

// DefaultRunStatusPublisher returns the production status publisher.
func DefaultRunStatusPublisher() RunStatusPublisher {
	return backupstatus.NewPublisher(backupstatus.PublisherDeps{})
}
