// Package config loads runtime configuration for cmd/api from a TOML file.
package config

import (
	"fmt"
	"net"
	"path/filepath"
	"time"
)

const (
	// DefaultConfigPath is the canonical host-admin configuration file.
	DefaultConfigPath = "/etc/vyntrio/config.toml"

	// CanonicalStateDir is the only approved persistent state directory in v1.
	CanonicalStateDir = "/var/lib/vyntrio"

	defaultVersion     = "0.2.0-dev"
	defaultBuildCommit = "unknown"

	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 15 * time.Second
)

// LoadOptions controls production versus test loading behavior.
type LoadOptions struct {
	// AllowedStateDir is the required exact state_dir value. When empty,
	// CanonicalStateDir is used.
	AllowedStateDir string
}

// Config holds API server runtime configuration.
type Config struct {
	BindAddress  string
	ListenPort   int
	StateDir     string
	LogLevel     string
	CookieSecure bool

	// Env is derived from cookie and bind settings for logging and settings DTO.
	Env             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Version         string
	BuildCommit     string
}

// Load reads and validates configuration from path using production defaults.
func Load(path string) (Config, error) {
	return LoadWithOptions(path, LoadOptions{})
}

// LoadWithOptions reads and validates configuration from path.
func LoadWithOptions(path string, opts LoadOptions) (Config, error) {
	allowedStateDir := opts.AllowedStateDir
	if allowedStateDir == "" {
		allowedStateDir = CanonicalStateDir
	}
	return loadFromFile(path, allowedStateDir)
}

// DatabasePath returns the SQLite database file path.
func (c Config) DatabasePath() string {
	return filepath.Join(c.StateDir, "vyntrio.db")
}

// Addr returns the host:port listen address.
func (c Config) Addr() string {
	return net.JoinHostPort(c.BindAddress, fmt.Sprintf("%d", c.ListenPort))
}
