package overview

import (
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

// Health summary labels for the read-only overview health slice.
const (
	HealthStatusHealthy = "healthy"
	HealthStatusWarning = "warning"
	HealthStatusUnknown = "unknown"

	HealthNoteDatabase = "database"
	HealthNoteBackup   = "backup"
)

// HealthSection exposes a coarse health summary derived from existing overview slices.
type HealthSection struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

// AssembleHealth maps already-materialized runtime and backup fields into the safe health summary.
func AssembleHealth(runtime RuntimeSection, backup backupstatus.Backup) HealthSection {
	switch runtime.Status {
	case RuntimeStatusUnknown:
		return HealthSection{Status: HealthStatusUnknown}
	case RuntimeStatusDegraded:
		if runtime.Note == RuntimeNoteDatabase {
			return HealthSection{Status: HealthStatusWarning, Note: HealthNoteDatabase}
		}
		return HealthSection{Status: HealthStatusWarning}
	case RuntimeStatusReady:
		if backup.Status == backupstatus.StatusFailed {
			return HealthSection{Status: HealthStatusWarning, Note: HealthNoteBackup}
		}
		return HealthSection{Status: HealthStatusHealthy}
	default:
		return HealthSection{Status: HealthStatusUnknown}
	}
}
