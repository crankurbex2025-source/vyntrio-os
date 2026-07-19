package config

import (
	"fmt"
	"os"
)

const (
	tlsCertFileKey = "tls_cert_file"
	tlsKeyFileKey  = "tls_key_file"
)

func validateTLSConfiguration(
	cookieSecure bool,
	bindAddress string,
	tlsCertFile string,
	tlsKeyFile string,
) error {
	hasCert := tlsCertFile != ""
	hasKey := tlsKeyFile != ""
	if hasCert != hasKey {
		return fmt.Errorf("config: tls_cert_file and tls_key_file must both be set or both omitted")
	}

	loopback := isLoopbackAddress(bindAddress)
	if !loopback {
		if !cookieSecure {
			return fmt.Errorf("config: cookie_secure=true required for non-loopback bind_address")
		}
		if !hasCert {
			return fmt.Errorf("config: tls_cert_file and tls_key_file required for non-loopback bind_address")
		}
	}

	if hasCert {
		if err := validateTLSFile(tlsCertFileKey, tlsCertFile); err != nil {
			return err
		}
		if err := validateTLSFile(tlsKeyFileKey, tlsKeyFile); err != nil {
			return err
		}
	}

	return nil
}

func validateTLSFile(key, path string) error {
	if path == "" {
		return fmt.Errorf("config: invalid value for %q", key)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("config: %s not readable: %s", key, path)
	}
	if info.IsDir() {
		return fmt.Errorf("config: %s must be a file: %s", key, path)
	}
	return nil
}
