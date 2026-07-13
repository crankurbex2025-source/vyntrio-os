package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func TestParseFlagsDefaultPath(t *testing.T) {
	path, err := config.ParseFlags(nil)
	if err != nil {
		t.Fatalf("ParseFlags() error: %v", err)
	}
	if path != config.DefaultConfigPath {
		t.Fatalf("path = %q, want %q", path, config.DefaultConfigPath)
	}
}

func TestParseFlagsConfigAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfigFile(t, dir, "config.toml", validConfigBody(testStateDir(t)))

	path, err := config.ParseFlags([]string{"--config", cfgPath})
	if err != nil {
		t.Fatalf("ParseFlags() error: %v", err)
	}
	if path != cfgPath {
		t.Fatalf("path = %q, want %q", path, cfgPath)
	}
}

func TestParseFlagsRejectsUnknownAndDuplicate(t *testing.T) {
	cfgPath := writeConfigFile(t, t.TempDir(), "config.toml", validConfigBody(testStateDir(t)))

	cases := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"--help"}},
		{name: "positional", args: []string{"extra"}},
		{name: "duplicate config", args: []string{"--config", cfgPath, "--config", cfgPath}},
		{name: "missing value", args: []string{"--config"}},
		{name: "empty value", args: []string{"--config", ""}},
		{name: "relative path", args: []string{"--config", "relative/config.toml"}},
		{name: "traversal path", args: []string{"--config", "/etc/../etc/vyntrio/config.toml"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := config.ParseFlags(tc.args); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestReadConfigRejectsSymlinkAndOversize(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("symlink test requires Unix")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "real.toml")
	if err := os.WriteFile(target, []byte(validConfigBody(testStateDir(t))), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.toml")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	linkAbs, err := filepath.Abs(link)
	if err != nil {
		t.Fatal(err)
	}
	stateDir := testStateDir(t)
	_, err = config.LoadWithOptions(linkAbs, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected symlink config file to fail")
	}

	huge := filepath.Join(dir, "huge.toml")
	if err := os.WriteFile(huge, []byte(strings.Repeat("a", 65*1024)), 0o600); err != nil {
		t.Fatal(err)
	}
	hugeAbs, err := filepath.Abs(huge)
	if err != nil {
		t.Fatal(err)
	}
	_, err = config.LoadWithOptions(hugeAbs, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected oversized config file to fail")
	}
}

func TestReadConfigRejectsWorldWritable(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("permission test requires Unix")
	}
	dir := t.TempDir()
	stateDir := testStateDir(t)
	path := writeConfigFile(t, dir, "config.toml", validConfigBody(stateDir))
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadWithOptions(path, config.LoadOptions{AllowedStateDir: stateDir})
	if err == nil {
		t.Fatal("expected world-writable config file to fail")
	}
}
