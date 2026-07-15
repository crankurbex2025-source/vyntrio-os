package overview_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
)

func TestAssembleRuntimeReady(t *testing.T) {
	got := overview.AssembleRuntime(
		overview.ReadinessSection{Status: "ready", Database: "ok"},
		overview.ServiceSection{Status: "running"},
	)
	if got.Status != overview.RuntimeStatusReady || got.Note != "" {
		t.Fatalf("runtime = %+v, want ready without note", got)
	}
}

func TestAssembleRuntimeDegradedDatabase(t *testing.T) {
	got := overview.AssembleRuntime(
		overview.ReadinessSection{Status: "not_ready", Database: "error"},
		overview.ServiceSection{Status: "running"},
	)
	if got.Status != overview.RuntimeStatusDegraded || got.Note != overview.RuntimeNoteDatabase {
		t.Fatalf("runtime = %+v, want degraded/database", got)
	}
}

func TestAssembleRuntimeUnknownWhenServiceNotRunning(t *testing.T) {
	got := overview.AssembleRuntime(
		overview.ReadinessSection{Status: "ready", Database: "ok"},
		overview.ServiceSection{Status: "stopped"},
	)
	if got.Status != overview.RuntimeStatusUnknown || got.Note != "" {
		t.Fatalf("runtime = %+v, want unknown without note", got)
	}
}

func TestAssembleRuntimeUnknownForUnexpectedReadinessShape(t *testing.T) {
	got := overview.AssembleRuntime(
		overview.ReadinessSection{Status: "ready", Database: "error"},
		overview.ServiceSection{Status: "running"},
	)
	if got.Status != overview.RuntimeStatusUnknown {
		t.Fatalf("runtime.status = %q, want unknown", got.Status)
	}
}
