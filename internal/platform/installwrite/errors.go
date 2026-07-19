package installwrite

import "errors"

var (
	ErrForceRequired         = errors.New("installwrite: --force required for mutation")
	ErrArtifactSourceRequired = errors.New("installwrite: envelope-root or release-manifest required")
	ErrUnsafeTargetRoot      = errors.New("installwrite: target root outside install sandbox")
	ErrPreflightFailed       = errors.New("installwrite: preflight failed")
	ErrPostVerifyFailed      = errors.New("installwrite: post-write verification failed")
	ErrInstallFailed         = errors.New("installwrite: install failed")
)
