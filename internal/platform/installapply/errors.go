package installapply

import "errors"

var (
	ErrForceRequired     = errors.New("installapply: --force required")
	ErrApplyFailed       = errors.New("installapply: partition apply failed")
	ErrPreflightFailed   = errors.New("installapply: preflight failed")
	ErrArtifactRequired  = errors.New("installapply: artifact source required")
	ErrUnsafeTargetRoot  = errors.New("installapply: target-root not allowed")
)
