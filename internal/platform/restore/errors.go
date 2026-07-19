package restore

import "errors"

var (
	ErrNotRoot                = errors.New("restore: must be run as root")
	ErrUsage                  = errors.New("restore: invalid usage")
	ErrForceRequired          = errors.New("restore: destructive restore requires --force")
	ErrArtifactInvalid        = errors.New("restore: artifact selection or preflight failed")
	ErrCompatibility          = errors.New("restore: compatibility gate failed")
	ErrServiceStopFailed      = errors.New("restore: failed to stop service")
	ErrServiceInactiveUnknown = errors.New("restore: service inactive state could not be proven")
	ErrServiceStartFailed     = errors.New("restore: service restart failed")
	ErrHealthProbeFailed      = errors.New("restore: local health probe failed")
	ErrReadinessProbeFailed   = errors.New("restore: local readiness probe failed")
	ErrPreservationFailed     = errors.New("restore: preservation copy failed")
	ErrPlacementFailed        = errors.New("restore: file placement failed")
	ErrOwnershipFailed        = errors.New("restore: ownership or mode repair failed")
	ErrRollbackFailed                 = errors.New("restore: rollback from preservation copy failed")
	ErrPostRestoreRollbackSucceeded   = errors.New("restore: post-restore failure; rolled back to preservation copy")
	ErrPostRestoreRollbackFailed      = errors.New("restore: post-restore failure; rollback also failed")
)

// FailureCategory classifies operator-visible restore failures.
type FailureCategory string

const (
	FailurePreflight     FailureCategory = "preflight"
	FailureCompatibility FailureCategory = "compatibility"
	FailureForceRequired FailureCategory = "force required"
	FailurePreservation  FailureCategory = "preservation"
	FailurePlacement     FailureCategory = "placement"
	FailureOwnership     FailureCategory = "ownership"
	FailureService       FailureCategory = "service"
	FailureHealth        FailureCategory = "health"
	FailureReadiness                 FailureCategory = "readiness"
	FailureRollback                  FailureCategory = "rollback"
	FailurePostRestoreRollbackOK     FailureCategory = "post-restore rollback succeeded"
	FailurePostRestoreRollbackFailed FailureCategory = "post-restore rollback failed"
)

func FailureClassFromError(err error) FailureCategory {
	switch {
	case errors.Is(err, ErrForceRequired):
		return FailureForceRequired
	case errors.Is(err, ErrCompatibility):
		return FailureCompatibility
	case errors.Is(err, ErrPreservationFailed):
		return FailurePreservation
	case errors.Is(err, ErrPlacementFailed):
		return FailurePlacement
	case errors.Is(err, ErrOwnershipFailed):
		return FailureOwnership
	case errors.Is(err, ErrServiceStopFailed), errors.Is(err, ErrServiceInactiveUnknown), errors.Is(err, ErrServiceStartFailed):
		return FailureService
	case errors.Is(err, ErrHealthProbeFailed):
		return FailureHealth
	case errors.Is(err, ErrReadinessProbeFailed):
		return FailureReadiness
	case errors.Is(err, ErrPostRestoreRollbackSucceeded):
		return FailurePostRestoreRollbackOK
	case errors.Is(err, ErrPostRestoreRollbackFailed):
		return FailurePostRestoreRollbackFailed
	case errors.Is(err, ErrRollbackFailed):
		return FailureRollback
	default:
		return FailurePreflight
	}
}
