package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var (
	destinationLstat    = os.Lstat
	destinationMkdir    = os.Mkdir
	destinationOwnerUID = destinationFileOwnerUID
)

func destinationFileOwnerUID(info os.FileInfo) (uint32, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return stat.Uid, true
}

// EnsureDestination prepares the fixed backup destination directory.
func EnsureDestination(path string) error {
	info, err := destinationLstat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("%w: destination inaccessible", ErrDestinationUnsafe)
		}
		if err := destinationMkdir(path, destinationMode); err != nil {
			return fmt.Errorf("%w: create destination", ErrDestinationUnsafe)
		}
		return verifyDestination(path)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: destination is symlink", ErrDestinationUnsafe)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: destination is not a directory", ErrDestinationUnsafe)
	}
	return verifyDestination(path)
}

func verifyDestination(path string) error {
	info, err := destinationLstat(path)
	if err != nil {
		return fmt.Errorf("%w: destination inaccessible", ErrDestinationUnsafe)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("%w: destination invalid", ErrDestinationUnsafe)
	}
	if info.Mode().Perm() != destinationMode {
		return fmt.Errorf("%w: destination mode must be 0700", ErrDestinationUnsafe)
	}
	uid, ok := destinationOwnerUID(info)
	if !ok || uid != 0 {
		return fmt.Errorf("%w: destination must be root-owned", ErrDestinationUnsafe)
	}
	return nil
}

func artifactPath(destination, timestamp string) string {
	return filepath.Join(destination, fmt.Sprintf("%s_%s.tar", FormatVersion, timestamp))
}

func tempArtifactPath(destination, timestamp string) string {
	return artifactPath(destination, timestamp) + ".tmp"
}

func finalizeArtifactMode(path string) error {
	return os.Chmod(path, artifactMode)
}
