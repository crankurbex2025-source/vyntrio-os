package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func testStateDir(t *testing.T) string {
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

func writeConfigFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Abs() error: %v", err)
	}
	return abs
}

func validConfigBody(stateDir string) string {
	return strings.TrimSpace(`
bind_address = "127.0.0.1"
listen_port = 8080
state_dir = "`+stateDir+`"
log_level = "info"
cookie_secure = false
`) + "\n"
}

func loadTestConfig(t *testing.T, body string) (config.Config, error) {
	t.Helper()
	stateDir := testStateDir(t)
	path := writeConfigFile(t, filepath.Dir(stateDir), "config.toml", strings.ReplaceAll(body, "__STATE_DIR__", stateDir))
	return config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
}

func TestLoadValidConfiguration(t *testing.T) {
	stateDir := testStateDir(t)
	path := writeConfigFile(t, t.TempDir(), "config.toml", validConfigBody(stateDir))

	cfg, err := config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
	if cfg.BindAddress != "127.0.0.1" || cfg.ListenPort != 8080 || cfg.StateDir != stateDir {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.LogLevel != "info" || cfg.CookieSecure {
		t.Fatalf("unexpected log/cookie config: %+v", cfg)
	}
	if cfg.Env != "development" {
		t.Fatalf("Env = %q, want development", cfg.Env)
	}
	if cfg.DatabasePath() != filepath.Join(stateDir, "vyntrio.db") {
		t.Fatalf("DatabasePath() = %q", cfg.DatabasePath())
	}
	if cfg.Addr() != "127.0.0.1:8080" {
		t.Fatalf("Addr() = %q", cfg.Addr())
	}
}

func TestLoadRejectsLegacyEnvironmentVariables(t *testing.T) {
	t.Setenv("VYNTRIO_API_HOST", "10.0.0.1")
	t.Setenv("VYNTRIO_API_PORT", "9090")
	t.Setenv("VYNTRIO_DATA_DIR", "/tmp/evil")
	t.Setenv("VYNTRIO_LOG_LEVEL", "debug")
	t.Setenv("VYNTRIO_ENV", "production")
	t.Setenv("VYNTRIO_COOKIE_SECURE", "true")

	stateDir := testStateDir(t)
	path := writeConfigFile(t, t.TempDir(), "config.toml", validConfigBody(stateDir))
	cfg, err := config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
	if cfg.BindAddress != "127.0.0.1" || cfg.ListenPort != 8080 || cfg.StateDir != stateDir || cfg.LogLevel != "info" || cfg.CookieSecure {
		t.Fatalf("environment variables affected config: %+v", cfg)
	}
}

func TestLoadMissingRequiredKey(t *testing.T) {
	body := strings.ReplaceAll(validConfigBody("__STATE_DIR__"), "log_level = \"info\"\n", "")
	_, err := loadTestConfig(t, body)
	if err == nil {
		t.Fatal("expected error for missing log_level")
	}
}

func TestLoadUnknownKey(t *testing.T) {
	body := validConfigBody("__STATE_DIR__") + "extra = true\n"
	_, err := loadTestConfig(t, body)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestLoadDuplicateKey(t *testing.T) {
	body := strings.Replace(validConfigBody("__STATE_DIR__"), "listen_port = 8080", "listen_port = 8080\nlisten_port = 9090", 1)
	_, err := loadTestConfig(t, body)
	if err == nil {
		t.Fatal("expected error for duplicate key")
	}
}

func TestLoadWrongTypes(t *testing.T) {
	cases := map[string]string{
		"bind_address":  `bind_address = 127`,
		"listen_port":   `listen_port = "8080"`,
		"state_dir":     `state_dir = 1`,
		"log_level":     `log_level = true`,
		"cookie_secure": `cookie_secure = "true"`,
	}
	for key, line := range cases {
		t.Run(key, func(t *testing.T) {
			body := validConfigBody("__STATE_DIR__")
			body = strings.Replace(body, strings.Split(line, " = ")[0]+" = "+strings.Split(strings.Split(line, " = ")[1], "\n")[0], line, 1)
			switch key {
			case "bind_address":
				body = strings.Replace(body, `bind_address = "127.0.0.1"`, line, 1)
			case "listen_port":
				body = strings.Replace(body, "listen_port = 8080", line, 1)
			case "state_dir":
				body = strings.Replace(body, `state_dir = "__STATE_DIR__"`, `state_dir = 1`, 1)
			case "log_level":
				body = strings.Replace(body, `log_level = "info"`, line, 1)
			case "cookie_secure":
				body = strings.Replace(body, "cookie_secure = false", line, 1)
			}
			_, err := loadTestConfig(t, body)
			if err == nil {
				t.Fatalf("expected error for wrong type on %s", key)
			}
		})
	}
}

func TestLoadNestedTableRejected(t *testing.T) {
	body := "[server]\nbind_address = \"127.0.0.1\"\n"
	_, err := loadTestConfig(t, body)
	if err == nil {
		t.Fatal("expected error for nested table config")
	}
}

func TestValidateBindAddress(t *testing.T) {
	loopbackCases := []string{"127.0.0.1", "::1"}
	for _, addr := range loopbackCases {
		body := strings.Replace(validConfigBody("__STATE_DIR__"), "127.0.0.1", addr, 1)
		if _, err := loadTestConfig(t, body); err != nil {
			t.Fatalf("valid bind_address %q failed: %v", addr, err)
		}
	}

	body := strings.ReplaceAll(validConfigBody("__STATE_DIR__"), "127.0.0.1", "192.168.1.10")
	body = strings.Replace(body, "cookie_secure = false", "cookie_secure = true", 1)
	if _, err := loadTestConfig(t, body); err != nil {
		t.Fatalf("valid non-loopback bind_address failed: %v", err)
	}

	invalid := []string{"localhost", "0.0.0.0", "::", "", " 127.0.0.1", "127.0.0.1/32", "http://127.0.0.1"}
	for _, addr := range invalid {
		body := strings.Replace(validConfigBody("__STATE_DIR__"), "127.0.0.1", addr, 1)
		if _, err := loadTestConfig(t, body); err == nil {
			t.Fatalf("invalid bind_address %q should fail", addr)
		}
	}
}

func TestValidateListenPort(t *testing.T) {
	for _, port := range []string{"0", "-1", "70000"} {
		body := strings.Replace(validConfigBody("__STATE_DIR__"), "listen_port = 8080", "listen_port = "+port, 1)
		if _, err := loadTestConfig(t, body); err == nil {
			t.Fatalf("port %s should fail", port)
		}
	}
}

func TestValidateLogLevel(t *testing.T) {
	for _, level := range []string{"INFO", " trace", "verbose"} {
		body := strings.Replace(validConfigBody("__STATE_DIR__"), "info", level, 1)
		if _, err := loadTestConfig(t, body); err == nil {
			t.Fatalf("log_level %q should fail", level)
		}
	}
}

func TestCookieSecureFalseRequiresLoopback(t *testing.T) {
	body := strings.ReplaceAll(validConfigBody("__STATE_DIR__"), "127.0.0.1", "192.168.1.10")
	_, err := loadTestConfig(t, body)
	if err == nil {
		t.Fatal("expected cookie_secure=false with non-loopback bind to fail")
	}
}

func TestCookieSecureTrueOnLoopback(t *testing.T) {
	body := strings.Replace(validConfigBody("__STATE_DIR__"), "cookie_secure = false", "cookie_secure = true", 1)
	cfg, err := loadTestConfig(t, body)
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
	if !cfg.CookieSecure || cfg.Env != "production" {
		t.Fatalf("unexpected secure config: %+v", cfg)
	}
}

func TestStateDirMustMatchAllowedRootExactly(t *testing.T) {
	stateDir := testStateDir(t)
	child := filepath.Join(filepath.Dir(stateDir), "child")
	if err := os.MkdirAll(child, 0o750); err != nil {
		t.Fatal(err)
	}
	path := writeConfigFile(t, t.TempDir(), "config.toml", validConfigBody(child))
	_, err := config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected mismatch between state_dir and allowed root to fail")
	}
}

func TestStateDirRejectsRelativePath(t *testing.T) {
	body := strings.Replace(validConfigBody("__STATE_DIR__"), "__STATE_DIR__", "relative/path", 1)
	_, err := loadTestConfig(t, body)
	if err == nil {
		t.Fatal("expected relative state_dir to fail")
	}
}

func TestStartupRejectsStateDirSymlink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	root := t.TempDir()
	realDir := filepath.Join(root, "real")
	linkDir := filepath.Join(root, "link")
	if err := os.MkdirAll(realDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}
	realAbs, err := filepath.Abs(realDir)
	if err != nil {
		t.Fatal(err)
	}
	linkAbs, err := filepath.Abs(linkDir)
	if err != nil {
		t.Fatal(err)
	}
	path := writeConfigFile(t, root, "config.toml", validConfigBody(linkAbs))
	_, err = config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: realAbs})
	if err == nil {
		t.Fatal("expected symlink state_dir to fail")
	}
}

func TestConfigErrorsDoNotExposeSecrets(t *testing.T) {
	body := `bind_address = "127.0.0.1"
listen_port = 8080
state_dir = "__STATE_DIR__"
log_level = "info"
cookie_secure = false
secret_token = "do-not-leak"
`
	errMsg := ""
	stateDir := testStateDir(t)
	path := writeConfigFile(t, t.TempDir(), "config.toml", strings.ReplaceAll(body, "__STATE_DIR__", stateDir))
	_, err := config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		errMsg = err.Error()
	}
	if strings.Contains(errMsg, "do-not-leak") {
		t.Fatalf("error leaked config content: %q", errMsg)
	}
}

func symlinkInStateDir(t *testing.T, stateDir, name, target string) {
	t.Helper()
	if err := os.Symlink(target, filepath.Join(stateDir, name)); err != nil {
		t.Fatalf("Symlink() error: %v", err)
	}
}

func loadConfigForStateDir(t *testing.T, stateDir string) error {
	t.Helper()
	path := writeConfigFile(t, filepath.Dir(stateDir), "config.toml", validConfigBody(stateDir))
	_, err := config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
	return err
}

func TestStartupRejectsExistingDatabaseSymlink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStateDir(t)
	symlinkInStateDir(t, stateDir, "vyntrio.db", t.TempDir())
	if err := loadConfigForStateDir(t, stateDir); err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db symlink")
	}
}

func TestStartupRejectsExistingJournalSymlink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStateDir(t)
	symlinkInStateDir(t, stateDir, "vyntrio.db-journal", t.TempDir())
	if err := loadConfigForStateDir(t, stateDir); err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db-journal symlink")
	}
}

func TestStartupRejectsExistingWALSymlink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStateDir(t)
	symlinkInStateDir(t, stateDir, "vyntrio.db-wal", t.TempDir())
	if err := loadConfigForStateDir(t, stateDir); err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db-wal symlink")
	}
}

func TestStartupRejectsExistingSHMSymlink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStateDir(t)
	symlinkInStateDir(t, stateDir, "vyntrio.db-shm", t.TempDir())
	if err := loadConfigForStateDir(t, stateDir); err == nil {
		t.Fatal("expected startup-time rejection of vyntrio.db-shm symlink")
	}
}

func TestStartupAcceptsMissingDatabaseAndSidecars(t *testing.T) {
	stateDir := testStateDir(t)
	if err := loadConfigForStateDir(t, stateDir); err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
}

func TestStartupAcceptsRegularExistingDatabaseFile(t *testing.T) {
	stateDir := testStateDir(t)
	dbPath := filepath.Join(stateDir, "vyntrio.db")
	if err := os.WriteFile(dbPath, []byte("not-a-valid-sqlite-header"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(dbPath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("expected regular database file before load, got err=%v mode=%v", err, info.Mode())
	}
	if err := loadConfigForStateDir(t, stateDir); err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
}

func TestStartupSQLiteSymlinkErrorsDoNotExposeTargets(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("startup-time symlink rejection requires Unix symlink support")
	}
	stateDir := testStateDir(t)
	secretTarget := filepath.Join(t.TempDir(), "secret-target")
	if err := os.MkdirAll(secretTarget, 0o700); err != nil {
		t.Fatal(err)
	}
	symlinkInStateDir(t, stateDir, "vyntrio.db", secretTarget)
	err := loadConfigForStateDir(t, stateDir)
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, secretTarget) {
		t.Fatalf("error exposed symlink target: %q", errMsg)
	}
}
