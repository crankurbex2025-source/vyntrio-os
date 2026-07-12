package config_test

import (
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("VYNTRIO_API_PORT", "")
	t.Setenv("VYNTRIO_API_HOST", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.APIHost != "127.0.0.1" {
		t.Errorf("APIHost = %q, want 127.0.0.1", cfg.APIHost)
	}
	if cfg.APIPort != 8080 {
		t.Errorf("APIPort = %d, want 8080", cfg.APIPort)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout = %v, want 15s", cfg.ReadTimeout)
	}
	if cfg.Version != "0.2.0-dev" {
		t.Errorf("Version = %q, want 0.2.0-dev", cfg.Version)
	}
}

func TestLoadCustomPort(t *testing.T) {
	t.Setenv("VYNTRIO_API_PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Addr() != "127.0.0.1:9090" {
		t.Errorf("Addr() = %q, want 127.0.0.1:9090", cfg.Addr())
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("VYNTRIO_API_PORT", "70000")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
