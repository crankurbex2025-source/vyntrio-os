package backup

import "errors"

var (
	ErrNotRoot                = errors.New("backup: must be run as root")
	ErrServiceNotActive       = errors.New("backup: service is not active")
	ErrServiceStopFailed      = errors.New("backup: failed to stop service")
	ErrServiceInactiveUnknown = errors.New("backup: service inactive state could not be proven")
	ErrServiceStartFailed     = errors.New("backup: service restart failed")
	ErrHealthProbeFailed      = errors.New("backup: local health probe failed")
	ErrReadinessProbeFailed   = errors.New("backup: local readiness probe failed")
	ErrArtifactFailed         = errors.New("backup: artifact build or validation failed")
	ErrDestinationUnsafe      = errors.New("backup: destination directory is unsafe")
	ErrSourceInvalid          = errors.New("backup: source file validation failed")
	ErrArtifactCollision      = errors.New("backup: artifact destination already exists")
	ErrStatusPublishFailed    = errors.New("backup: status publication failed")
)

// FailureCategory classifies operator-visible backup failures.
type FailureCategory string

const (
	FailureArtifact  FailureCategory = "backup artifact failure"
	FailureRestart   FailureCategory = "service restart failure"
	FailureHealth    FailureCategory = "health probe failure"
	FailureReadiness FailureCategory = "readiness probe failure"
)
