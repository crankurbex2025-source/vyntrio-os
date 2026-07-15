package overview_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

func TestAssembleHealthHealthy(t *testing.T) {
	got := overview.AssembleHealth(
		overview.RuntimeSection{Status: overview.RuntimeStatusReady},
		backupstatus.NeverRun(),
	)
	if got.Status != overview.HealthStatusHealthy || got.Note != "" {
		t.Fatalf("health = %+v, want healthy without note", got)
	}
}

func TestAssembleHealthWarningDatabase(t *testing.T) {
	got := overview.AssembleHealth(
		overview.RuntimeSection{Status: overview.RuntimeStatusDegraded, Note: overview.RuntimeNoteDatabase},
		backupstatus.NeverRun(),
	)
	if got.Status != overview.HealthStatusWarning || got.Note != overview.HealthNoteDatabase {
		t.Fatalf("health = %+v, want warning/database", got)
	}
}

func TestAssembleHealthWarningBackup(t *testing.T) {
	got := overview.AssembleHealth(
		overview.RuntimeSection{Status: overview.RuntimeStatusReady},
		backupstatus.Backup{Status: backupstatus.StatusFailed},
	)
	if got.Status != overview.HealthStatusWarning || got.Note != overview.HealthNoteBackup {
		t.Fatalf("health = %+v, want warning/backup", got)
	}
}

func TestAssembleHealthUnknown(t *testing.T) {
	got := overview.AssembleHealth(
		overview.RuntimeSection{Status: overview.RuntimeStatusUnknown},
		backupstatus.NeverRun(),
	)
	if got.Status != overview.HealthStatusUnknown || got.Note != "" {
		t.Fatalf("health = %+v, want unknown without note", got)
	}
}
