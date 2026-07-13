package main_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func writeStartupConfig(t *testing.T, stateDir string) string {
	t.Helper()
	body := strings.TrimSpace(`
bind_address = "127.0.0.1"
listen_port = 18080
state_dir = "`+stateDir+`"
log_level = "info"
cookie_secure = false
`) + "\n"
	path := filepath.Join(filepath.Dir(stateDir), "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Abs() error: %v", err)
	}
	return abs
}

func testStartupStateDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "var", "lib", "vyntrio")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("Abs() error: %v", err)
	}
	return abs
}

func TestAPIStartupLoadsConfigAndDatabase(t *testing.T) {
	stateDir := testStartupStateDir(t)
	configPath := writeStartupConfig(t, stateDir)

	parsedPath, err := config.ParseFlags([]string{"--config", configPath})
	if err != nil {
		t.Fatalf("ParseFlags() error: %v", err)
	}

	cfg, err := config.LoadWithOptions(parsedPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
	if cfg.Addr() != "127.0.0.1:18080" {
		t.Fatalf("Addr() = %q", cfg.Addr())
	}

	store, err := sqlite.Open(context.Background(), cfg.StateDir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if store.Path() != cfg.DatabasePath() {
		t.Fatalf("database path = %q, want %q", store.Path(), cfg.DatabasePath())
	}
}

func TestAPIStartupRejectsDatabaseSymlinkBeforeSQLiteOpen(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStartupStateDir(t)
	if err := os.Symlink(t.TempDir(), filepath.Join(stateDir, "vyntrio.db")); err != nil {
		t.Fatalf("Symlink() error: %v", err)
	}
	configPath := writeStartupConfig(t, stateDir)

	_, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db symlink")
	}

	info, err := os.Lstat(filepath.Join(stateDir, "vyntrio.db"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected vyntrio.db symlink to remain unchanged after rejected startup")
	}
}

func TestAPIStartupRejectsJournalSymlinkBeforeSQLiteOpen(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStartupStateDir(t)
	if err := os.Symlink(t.TempDir(), filepath.Join(stateDir, "vyntrio.db-journal")); err != nil {
		t.Fatalf("Symlink() error: %v", err)
	}
	configPath := writeStartupConfig(t, stateDir)

	_, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db-journal symlink")
	}
}

func TestAPIStartupRejectsWALSymlinkBeforeSQLiteOpen(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStartupStateDir(t)
	if err := os.Symlink(t.TempDir(), filepath.Join(stateDir, "vyntrio.db-wal")); err != nil {
		t.Fatalf("Symlink() error: %v", err)
	}
	configPath := writeStartupConfig(t, stateDir)

	_, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db-wal symlink")
	}
}

func TestAPIStartupRejectsSHMSymlinkBeforeSQLiteOpen(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStartupStateDir(t)
	if err := os.Symlink(t.TempDir(), filepath.Join(stateDir, "vyntrio.db-shm")); err != nil {
		t.Fatalf("Symlink() error: %v", err)
	}
	configPath := writeStartupConfig(t, stateDir)

	_, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db-shm symlink")
	}
}

func TestAPIStartupAcceptsMissingDatabaseAndSidecars(t *testing.T) {
	stateDir := testStartupStateDir(t)
	configPath := writeStartupConfig(t, stateDir)

	cfg, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}

	store, err := sqlite.Open(context.Background(), cfg.StateDir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
}

func TestAPIStartupAcceptsRegularExistingDatabaseFile(t *testing.T) {
	stateDir := testStartupStateDir(t)
	configPath := writeStartupConfig(t, stateDir)

	cfg, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}

	store, err := sqlite.Open(context.Background(), cfg.StateDir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	if _, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir}); err != nil {
		t.Fatalf("restart LoadWithOptions() error: %v", err)
	}
	store, err = sqlite.Open(context.Background(), cfg.StateDir)
	if err != nil {
		t.Fatalf("restart Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
}

func TestAPIStartupInvalidConfigDoesNotOpenDatabase(t *testing.T) {
	stateDir := testStartupStateDir(t)
	configPath := writeStartupConfig(t, stateDir)
	body, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	invalid := strings.Replace(string(body), "listen_port = 18080", "listen_port = 0", 1)
	if err := os.WriteFile(configPath, []byte(invalid), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err = config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected invalid config to fail")
	}

	if _, err := os.Stat(filepath.Join(stateDir, "vyntrio.db")); err == nil {
		t.Fatal("database file should not be created for invalid config")
	}
}
