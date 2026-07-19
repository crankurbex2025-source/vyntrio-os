package restore

import (
	"fmt"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

// CheckCompatibility applies fail-closed restore compatibility rules before mutation.
func CheckCompatibility(manifest backup.Manifest) error {
	if manifest.FormatVersion != backup.FormatVersion {
		return fmt.Errorf("%w: unsupported format %q", ErrCompatibility, manifest.FormatVersion)
	}
	if manifest.MigrationDirty != nil && *manifest.MigrationDirty {
		return fmt.Errorf("%w: backup recorded dirty migration state", ErrCompatibility)
	}
	return nil
}

// ScopeSummary returns a stable operator-facing restore scope description.
func ScopeSummary(manifest backup.Manifest) string {
	hasConfig := false
	stateMembers := 0
	for _, member := range manifest.Members {
		switch member.Name {
		case backup.ConfigMember:
			hasConfig = true
		case backup.StateDBMember, backup.StateJournalMem, backup.StateWALMember, backup.StateSHMMember:
			stateMembers++
		}
	}
	scope := "sqlite-state"
	if hasConfig {
		scope += ",config"
	}
	return fmt.Sprintf("%s members=%d", scope, len(manifest.Members))
}
