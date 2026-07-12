// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds API server configuration for cmd/api.
type Config struct {
	Env          string
	LogLevel     string
	APIHost      string
	APIPort      int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Version      string
	BuildCommit  string
	DataDir      string
}

// Load reads configuration from environment variables with documented defaults.
func Load() (Config, error) {
	port, err := strconv.Atoi(getEnv("VYNTRIO_API_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("VYNTRIO_API_PORT: %w", err)
	}

	readTimeout, err := time.ParseDuration(getEnv("VYNTRIO_API_READ_TIMEOUT", "15s"))
	if err != nil {
		return Config{}, fmt.Errorf("VYNTRIO_API_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := time.ParseDuration(getEnv("VYNTRIO_API_WRITE_TIMEOUT", "15s"))
	if err != nil {
		return Config{}, fmt.Errorf("VYNTRIO_API_WRITE_TIMEOUT: %w", err)
	}

	idleTimeout, err := time.ParseDuration(getEnv("VYNTRIO_API_IDLE_TIMEOUT", "60s"))
	if err != nil {
		return Config{}, fmt.Errorf("VYNTRIO_API_IDLE_TIMEOUT: %w", err)
	}

	cfg := Config{
		Env:          getEnv("VYNTRIO_ENV", "development"),
		LogLevel:     getEnv("VYNTRIO_LOG_LEVEL", "info"),
		APIHost:      getEnv("VYNTRIO_API_HOST", "127.0.0.1"),
		APIPort:      port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Version:      getEnv("VYNTRIO_VERSION", "0.2.0-dev"),
		BuildCommit:  getEnv("VYNTRIO_BUILD_COMMIT", "unknown"),
		DataDir:      getEnv("VYNTRIO_DATA_DIR", "./data"),
	}

	if cfg.APIPort < 1 || cfg.APIPort > 65535 {
		return Config{}, fmt.Errorf("VYNTRIO_API_PORT: must be 1-65535")
	}

	return cfg, nil
}

// DatabasePath returns the SQLite database file path.
func (c Config) DatabasePath() string {
	return filepath.Join(c.DataDir, "vyntrio.db")
}

// Addr returns the host:port listen address.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.APIHost, c.APIPort)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
