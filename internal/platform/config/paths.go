package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const maxConfigFileSize = 64 * 1024

func validateConfigFilePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("config: configuration path must be absolute")
	}
	if hasTraversalComponent(path) {
		return fmt.Errorf("config: configuration path is invalid")
	}
	clean := filepath.Clean(path)
	if clean != path {
		return fmt.Errorf("config: configuration path is invalid")
	}
	return nil
}

func readConfigFile(path string) ([]byte, error) {
	if err := validateConfigFilePath(path); err != nil {
		return nil, err
	}

	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config: configuration file not found: %s", path)
		}
		return nil, fmt.Errorf("config: configuration file is not accessible: %s", path)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("config: configuration file must be a regular file: %s", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("config: configuration file must be a regular file: %s", path)
	}
	if info.Size() > maxConfigFileSize {
		return nil, fmt.Errorf("config: configuration file exceeds size limit: %s", path)
	}
	if err := rejectWorldWritable(path, info); err != nil {
		return nil, err
	}
	if err := rejectWorldWritableDir(filepath.Dir(path)); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: configuration file is not accessible: %s", path)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("config: configuration file is invalid: %s", path)
	}
	return data, nil
}

func rejectWorldWritable(path string, info os.FileInfo) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if info.Mode().Perm()&0o002 != 0 {
		return fmt.Errorf("config: configuration file must not be world-writable: %s", path)
	}
	return nil
}

func rejectWorldWritableDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf("config: configuration directory is not accessible: %s", dir)
	}
	if info.Mode().Perm()&0o002 != 0 {
		return fmt.Errorf("config: configuration directory must not be world-writable: %s", dir)
	}
	return nil
}

func validateStateDir(stateDir, allowedStateDir string) (string, error) {
	if !filepath.IsAbs(stateDir) {
		return "", fmt.Errorf("config: state_dir must be an absolute path")
	}
	if strings.TrimSpace(stateDir) == "" {
		return "", fmt.Errorf("config: state_dir must be an absolute path")
	}
	if hasTraversalComponent(stateDir) {
		return "", fmt.Errorf("config: state_dir is invalid")
	}

	clean := filepath.Clean(stateDir)
	if clean != stateDir {
		return "", fmt.Errorf("config: state_dir is invalid")
	}
	if clean != allowedStateDir {
		return "", fmt.Errorf("config: state_dir must equal %s", allowedStateDir)
	}

	if err := os.MkdirAll(clean, 0o750); err != nil {
		return "", fmt.Errorf("config: state directory is not accessible")
	}

	info, err := os.Lstat(clean)
	if err != nil {
		return "", fmt.Errorf("config: state directory is not accessible")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("config: state_dir must not be a symlink")
	}

	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", fmt.Errorf("config: state directory is not accessible")
	}
	if resolved != allowedStateDir {
		return "", fmt.Errorf("config: state_dir resolves outside the allowed state root")
	}

	root, err := os.OpenRoot(resolved)
	if err != nil {
		return "", fmt.Errorf("config: state directory is not accessible")
	}
	defer root.Close() // startup accessibility probe only; does not bind SQLite I/O

	if err := rejectExistingSQLiteSymlinks(resolved); err != nil {
		return "", err
	}

	return resolved, nil
}

var sqliteStateFilenames = []string{
	"vyntrio.db",
	"vyntrio.db-journal",
	"vyntrio.db-wal",
	"vyntrio.db-shm",
}

// rejectExistingSQLiteSymlinks performs startup-time symlink rejection for known
// SQLite main-DB and sidecar names. It does not provide race-free protection
// against post-check filesystem mutation.
func rejectExistingSQLiteSymlinks(stateDir string) error {
	for _, name := range sqliteStateFilenames {
		path := filepath.Join(stateDir, name)
		info, err := os.Lstat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("config: state directory is not accessible")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("config: state directory contains an invalid database file")
		}
	}
	return nil
}

func hasTraversalComponent(path string) bool {
	slash := filepath.ToSlash(path)
	for _, part := range strings.Split(slash, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}
