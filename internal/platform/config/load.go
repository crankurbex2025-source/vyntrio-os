package config

import (
	"bytes"
	"fmt"
	"net/netip"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var allowedLogLevels = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
}

var requiredConfigKeys = []string{
	"bind_address",
	"listen_port",
	"state_dir",
	"log_level",
	"cookie_secure",
}

func loadFromFile(path, allowedStateDir string) (Config, error) {
	data, err := readConfigFile(path)
	if err != nil {
		return Config{}, err
	}

	raw, err := decodeStrictTOML(data)
	if err != nil {
		return Config{}, fmt.Errorf("config: invalid configuration file %s", path)
	}

	bindAddress, err := requiredString(raw, "bind_address")
	if err != nil {
		return Config{}, err
	}
	if err := validateBindAddress(bindAddress); err != nil {
		return Config{}, err
	}

	listenPort, err := requiredInt(raw, "listen_port")
	if err != nil {
		return Config{}, err
	}
	if err := validateListenPort(listenPort); err != nil {
		return Config{}, err
	}

	stateDirValue, err := requiredString(raw, "state_dir")
	if err != nil {
		return Config{}, err
	}
	stateDir, err := validateStateDir(stateDirValue, allowedStateDir)
	if err != nil {
		return Config{}, err
	}

	logLevel, err := requiredString(raw, "log_level")
	if err != nil {
		return Config{}, err
	}
	if err := validateLogLevel(logLevel); err != nil {
		return Config{}, err
	}

	cookieSecure, err := requiredBool(raw, "cookie_secure")
	if err != nil {
		return Config{}, err
	}
	if err := validateCookieSecure(cookieSecure, bindAddress); err != nil {
		return Config{}, err
	}

	env := "production"
	if !cookieSecure && isLoopbackAddress(bindAddress) {
		env = "development"
	}

	return Config{
		BindAddress:     bindAddress,
		ListenPort:      listenPort,
		StateDir:        stateDir,
		LogLevel:        logLevel,
		CookieSecure:    cookieSecure,
		Env:             env,
		ReadTimeout:     defaultReadTimeout,
		WriteTimeout:    defaultWriteTimeout,
		IdleTimeout:     defaultIdleTimeout,
		ShutdownTimeout: defaultShutdownTimeout,
		Version:         defaultVersion,
		BuildCommit:     defaultBuildCommit,
	}, nil
}

func decodeStrictTOML(data []byte) (map[string]any, error) {
	dec := toml.NewDecoder(bytes.NewReader(data))
	// Strict five-key schema is enforced below by required-key presence, exact
	// key count, and per-field type checks. DisallowUnknownFields applies only
	// to struct targets and is not used here.

	var raw map[string]any
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("config: empty configuration")
	}

	allowed := make(map[string]struct{}, len(requiredConfigKeys))
	for _, key := range requiredConfigKeys {
		allowed[key] = struct{}{}
	}

	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return nil, fmt.Errorf("config: unknown key %q", key)
		}
	}
	for _, key := range requiredConfigKeys {
		if _, ok := raw[key]; !ok {
			return nil, fmt.Errorf("config: missing required key %q", key)
		}
	}
	if len(raw) != len(requiredConfigKeys) {
		return nil, fmt.Errorf("config: invalid configuration schema")
	}

	return raw, nil
}

func requiredString(raw map[string]any, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", fmt.Errorf("config: missing required key %q", key)
	}
	typed, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("config: invalid type for %q", key)
	}
	if strings.TrimSpace(typed) != typed || typed == "" {
		return "", fmt.Errorf("config: invalid value for %q", key)
	}
	return typed, nil
}

func requiredInt(raw map[string]any, key string) (int, error) {
	value, ok := raw[key]
	if !ok {
		return 0, fmt.Errorf("config: missing required key %q", key)
	}
	switch typed := value.(type) {
	case int64:
		return int(typed), nil
	case int:
		return typed, nil
	default:
		return 0, fmt.Errorf("config: invalid type for %q", key)
	}
}

func requiredBool(raw map[string]any, key string) (bool, error) {
	value, ok := raw[key]
	if !ok {
		return false, fmt.Errorf("config: missing required key %q", key)
	}
	typed, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("config: invalid type for %q", key)
	}
	return typed, nil
}

func validateBindAddress(value string) error {
	if strings.Contains(value, "%") {
		return fmt.Errorf("config: invalid bind_address")
	}
	addr, err := netip.ParseAddr(value)
	if err != nil || !addr.IsValid() {
		return fmt.Errorf("config: invalid bind_address")
	}
	if addr.IsUnspecified() {
		return fmt.Errorf("config: invalid bind_address")
	}
	return nil
}

func validateListenPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("config: invalid listen_port")
	}
	return nil
}

func validateLogLevel(value string) error {
	if _, ok := allowedLogLevels[value]; !ok {
		return fmt.Errorf("config: invalid log_level")
	}
	return nil
}

func validateCookieSecure(cookieSecure bool, bindAddress string) error {
	if cookieSecure {
		return nil
	}
	if !isLoopbackAddress(bindAddress) {
		return fmt.Errorf("config: cookie_secure=false requires loopback bind_address")
	}
	return nil
}

func isLoopbackAddress(value string) bool {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return false
	}
	return addr.IsLoopback()
}
