package identity

import (
	"errors"
	"strings"
	"unicode"
)

const maxBootstrapUsernameBytes = 64

// ErrInvalidBootstrapUsername indicates the bootstrap username is invalid.
var ErrInvalidBootstrapUsername = errors.New("invalid bootstrap username")

// ValidateBootstrapUsername checks the smallest bootstrap username rules.
func ValidateBootstrapUsername(username string) error {
	if username == "" {
		return ErrInvalidBootstrapUsername
	}
	if len(username) > maxBootstrapUsernameBytes {
		return ErrInvalidBootstrapUsername
	}
	if strings.TrimSpace(username) != username {
		return ErrInvalidBootstrapUsername
	}
	for _, r := range username {
		if unicode.IsControl(r) {
			return ErrInvalidBootstrapUsername
		}
		if unicode.IsSpace(r) {
			return ErrInvalidBootstrapUsername
		}
	}
	return nil
}
