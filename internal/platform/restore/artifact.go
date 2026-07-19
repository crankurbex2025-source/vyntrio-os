package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

var completedArtifactPattern = regexp.MustCompile(
	`^` + regexp.QuoteMeta(backup.FormatVersion) + `_\d{8}T\d{6}\.\d{9}Z\.tar$`,
)

// ResolveArtifactPath maps an operator-supplied artifact basename to a fixed
// destination path. Arbitrary paths and incomplete artifacts are rejected.
func ResolveArtifactPath(basename, destinationDir string) (string, error) {
	if err := validateArtifactBasename(basename); err != nil {
		return "", err
	}
	if destinationDir == "" {
		destinationDir = DestinationDir
	}
	cleanDest := filepath.Clean(destinationDir)
	if cleanDest != destinationDir || strings.Contains(basename, string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: destination path invalid", ErrArtifactInvalid)
	}
	fullPath := filepath.Join(cleanDest, basename)
	if filepath.Dir(fullPath) != cleanDest {
		return "", fmt.Errorf("%w: artifact escapes destination", ErrArtifactInvalid)
	}
	info, err := os.Lstat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: artifact not found", ErrArtifactInvalid)
		}
		return "", fmt.Errorf("%w: artifact inaccessible", ErrArtifactInvalid)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%w: artifact is symlink", ErrArtifactInvalid)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%w: artifact is not a regular file", ErrArtifactInvalid)
	}
	return fullPath, nil
}

func validateArtifactBasename(basename string) error {
	if basename == "" || basename != filepath.Base(basename) {
		return fmt.Errorf("%w: malformed artifact name", ErrArtifactInvalid)
	}
	if strings.Contains(basename, "..") || strings.ContainsRune(basename, os.PathSeparator) {
		return fmt.Errorf("%w: traversal rejected", ErrArtifactInvalid)
	}
	if strings.HasSuffix(basename, ".tmp") {
		return fmt.Errorf("%w: incomplete artifact rejected", ErrArtifactInvalid)
	}
	if !completedArtifactPattern.MatchString(basename) {
		return fmt.Errorf("%w: artifact name does not match publication convention", ErrArtifactInvalid)
	}
	return nil
}
