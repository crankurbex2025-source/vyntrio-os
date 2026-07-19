package installtarget

import "errors"

var (
	ErrForceRequired            = errors.New("installtarget: --force required")
	ErrTargetNotEligible        = errors.New("installtarget: target not eligible")
	ErrUnsupportedTargetState   = errors.New("installtarget: unsupported target state")
	ErrAmbiguousTargetLayout    = errors.New("installtarget: ambiguous target layout")
	ErrTargetMounted            = errors.New("installtarget: target is mounted")
	ErrMountFailed              = errors.New("installtarget: mount failed")
	ErrApplyFailed              = errors.New("installtarget: apply failed")
	ErrRollbackFailed           = errors.New("installtarget: rollback failed")
)
