// Package backup implements the root-only local appliance backup command contract.
package backup

import "time"

const (
	FormatVersion = "vyntrio-backup-v1"

	ServiceName    = "vyntrio-api.service"
	StateRoot      = "/var/lib/vyntrio"
	ConfigPath     = "/etc/vyntrio/config.toml"
	DestinationDir = "/var/lib/vyntrio/backups"
	HealthBaseURL  = "http://127.0.0.1:8080"
	VersionURL     = HealthBaseURL + "/api/v1/version"

	MainDBName       = "vyntrio.db"
	JournalSidecar   = "vyntrio.db-journal"
	WALSidecar       = "vyntrio.db-wal"
	SHMSidecar       = "vyntrio.db-shm"
	ManifestFileName = "manifest.json"
	ConfigMember     = "config/config.toml"
	StateDBMember    = "state/vyntrio.db"
	StateJournalMem  = "state/vyntrio.db-journal"
	StateWALMember   = "state/vyntrio.db-wal"
	StateSHMMember   = "state/vyntrio.db-shm"

	destinationMode = 0o700
	artifactMode    = 0o600
	tarMemberMode   = 0o640

	defaultServiceTimeout = 30 * time.Second
	defaultHTTPTimeout    = 10 * time.Second

	// Post-start loopback verification tolerates the short gap between systemd
	// reporting active and the API listener accepting HTTP (~100 ms observed).
	postStartProbeDeadline      = 15 * time.Second
	postStartProbeInitialDelay  = 50 * time.Millisecond
	postStartProbeRetryInterval = 100 * time.Millisecond
)

// AllowedArchiveMembers is the fixed member set for vyntrio-backup-v1.
var AllowedArchiveMembers = map[string]struct{}{
	StateDBMember:    {},
	StateJournalMem:  {},
	StateWALMember:   {},
	StateSHMMember:   {},
	ConfigMember:     {},
	ManifestFileName: {},
}
