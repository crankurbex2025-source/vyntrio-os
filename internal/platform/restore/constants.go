// Package restore implements the root-only offline restore CLI contract (Block 7).
package restore

import (
	"os"
	"time"
)

const (
	FormatVersion = "vyntrio-backup-v1"

	ServiceName    = "vyntrio-api.service"
	StateRoot      = "/var/lib/vyntrio"
	ConfigPath     = "/etc/vyntrio/config.toml"
	DestinationDir = "/var/lib/vyntrio/backups"

	StateServiceAccount = "vyntrio"
	ConfigGroup         = "vyntrio"

	stateDirMode   os.FileMode = 0o750
	stateFileMode  os.FileMode = 0o640
	configFileMode os.FileMode = 0o640
	preservePrefix = "preserve-v1_"

	defaultServiceTimeout = 30 * time.Second
)

// ArchiveMemberToHost maps approved archive members to fixed host targets.
var ArchiveMemberToHost = map[string]string{
	"state/vyntrio.db":         StateRoot + "/vyntrio.db",
	"state/vyntrio.db-journal": StateRoot + "/vyntrio.db-journal",
	"state/vyntrio.db-wal":     StateRoot + "/vyntrio.db-wal",
	"state/vyntrio.db-shm":     StateRoot + "/vyntrio.db-shm",
	"config/config.toml":       ConfigPath,
}

// RestorableMembers lists archive paths that may replace live files.
var RestorableMembers = []string{
	"state/vyntrio.db",
	"state/vyntrio.db-journal",
	"state/vyntrio.db-wal",
	"state/vyntrio.db-shm",
	"config/config.toml",
}
