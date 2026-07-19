package installwrite

import "os"

const (
	DefaultSandboxRoot = "distro/installer/target-sandbox"
	InstallRecordName  = "INSTALL_RECORD.txt"
	SchemaVersion      = "vyntrio-installer-payload-install-v1"
)

var stateDirectories = []string{
	"etc/vyntrio",
	"var/lib/vyntrio",
	"var/lib/vyntrio/backups",
}

var payloadModes = map[string]os.FileMode{
	"usr/bin/vyntrio-api":                          0o755,
	"usr/bin/vyntrio-backup":                       0o755,
	"etc/systemd/system/vyntrio-api.service":       0o644,
	"usr/lib/sysusers.d/vyntrio.conf":              0o644,
	"etc/tmpfiles.d/vyntrio.conf":                  0o644,
	"etc/vyntrio/config.toml":                      0o640,
}
