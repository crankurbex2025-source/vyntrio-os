package mediacreator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// StartupLogPath returns a writable log path for GUI diagnostics.
func StartupLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "vyntrio-media-creator.log")
	}
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Local")
		}
		dir := filepath.Join(base, "Vyntrio", "media-creator")
		_ = os.MkdirAll(dir, 0o755)
		return filepath.Join(dir, "media-creator.log")
	case "darwin":
		dir := filepath.Join(home, "Library", "Logs", "Vyntrio")
		_ = os.MkdirAll(dir, 0o755)
		return filepath.Join(dir, "media-creator.log")
	default:
		dir := filepath.Join(home, ".local", "state", "vyntrio")
		_ = os.MkdirAll(dir, 0o755)
		return filepath.Join(dir, "media-creator.log")
	}
}

// AppendLog writes a timestamped diagnostic line.
func AppendLog(path, line string) {
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "%s %s\n", time.Now().UTC().Format(time.RFC3339), line)
}
