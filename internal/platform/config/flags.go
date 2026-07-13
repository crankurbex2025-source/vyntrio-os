package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParseFlags parses startup-only CLI arguments and returns the configuration
// file path. Only --config is supported.
func ParseFlags(args []string) (string, error) {
	configPath := DefaultConfigPath
	var configSeen bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--config":
			if configSeen {
				return "", fmt.Errorf("config: duplicate --config flag")
			}
			configSeen = true
			if i+1 >= len(args) {
				return "", fmt.Errorf("config: --config requires a value")
			}
			value := args[i+1]
			if strings.TrimSpace(value) == "" {
				return "", fmt.Errorf("config: --config requires a value")
			}
			if err := validateConfigFilePath(value); err != nil {
				return "", err
			}
			configPath = filepath.Clean(value)
			i++
		case strings.HasPrefix(arg, "-"):
			return "", fmt.Errorf("config: unknown flag %q", arg)
		default:
			return "", fmt.Errorf("config: unexpected argument %q", arg)
		}
	}

	return configPath, nil
}
