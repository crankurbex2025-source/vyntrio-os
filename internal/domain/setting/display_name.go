package setting

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxInstanceDisplayNameRunes = 80

// ValidateInstanceDisplayName validates and returns the canonical trimmed instance display name.
func ValidateInstanceDisplayName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%w: display name must not be empty", ErrInvalidValue)
	}
	if utf8.RuneCountInString(trimmed) > MaxInstanceDisplayNameRunes {
		return "", fmt.Errorf("%w: display name exceeds maximum length", ErrInvalidValue)
	}
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("%w: display name contains invalid characters", ErrInvalidValue)
		}
	}
	return trimmed, nil
}
