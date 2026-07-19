package releaseartifact

import "errors"

var (
	ErrMalformedManifest = errors.New("releaseartifact: malformed manifest")
	ErrVerifyFailed      = errors.New("releaseartifact: verification failed")
)
