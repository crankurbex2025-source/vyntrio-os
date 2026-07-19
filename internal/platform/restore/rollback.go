package restore

import (
	"context"
	"errors"
	"fmt"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

type postPlacementContext struct {
	result      Result
	preserveDir string
	opts        Options
	members     map[string]struct{}
	targets     OwnershipTargets
}

// completeRestore starts the service and runs local health/readiness probes.
func (r *Runner) completeRestore(ctx context.Context) error {
	if err := r.Service.Start(ctx); err != nil {
		return ErrServiceStartFailed
	}
	active, err := r.Service.IsActive(ctx)
	if err != nil || !active {
		return ErrServiceStartFailed
	}
	if err := r.Health.Probe(ctx); err != nil {
		if errors.Is(err, backup.ErrReadinessProbeFailed) {
			return ErrReadinessProbeFailed
		}
		return ErrHealthProbeFailed
	}
	return nil
}

// handlePostPlacementFailure rolls back restored files from the preservation copy,
// repairs ownership, restarts the service, and reprobes when practical.
func (r *Runner) handlePostPlacementFailure(
	ctx context.Context,
	fc postPlacementContext,
	cause error,
) (Result, error) {
	result := fc.result
	result.RollbackAttempted = true

	if err := r.Service.Stop(ctx); err != nil {
		return result, fmt.Errorf("%w: stop before rollback: %v", ErrPostRestoreRollbackFailed, cause)
	}

	if err := rollbackFromPreserve(fc.preserveDir, fc.opts, fc.members); err != nil {
		return result, fmt.Errorf("%w: %v", ErrPostRestoreRollbackFailed, err)
	}
	if err := applyOwnership(fc.targets, fc.opts.Ownership); err != nil {
		return result, fmt.Errorf("%w: %v", ErrPostRestoreRollbackFailed, err)
	}
	result.RollbackSucceeded = true

	if err := r.Service.Start(ctx); err != nil {
		return result, fmt.Errorf("%w: restart after rollback: %v", ErrPostRestoreRollbackFailed, err)
	}
	active, err := r.Service.IsActive(ctx)
	if err != nil || !active {
		return result, fmt.Errorf("%w: service inactive after rollback restart: %v", ErrPostRestoreRollbackFailed, cause)
	}

	if probeErr := r.Health.Probe(ctx); probeErr != nil {
		return result, fmt.Errorf("%w: %v", ErrPostRestoreRollbackSucceeded, cause)
	}
	return result, fmt.Errorf("%w: %v", ErrPostRestoreRollbackSucceeded, cause)
}
