package overview

// Runtime status labels for the read-only runtime-readiness overview slice.
const (
	RuntimeStatusReady    = "ready"
	RuntimeStatusDegraded = "degraded"
	RuntimeStatusUnknown  = "unknown"

	RuntimeNoteDatabase = "database"
)

// RuntimeSection exposes a coarse runtime-readiness label derived from existing overview state.
type RuntimeSection struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

// AssembleRuntime maps already-materialized readiness and service fields into the safe runtime slice.
func AssembleRuntime(readiness ReadinessSection, service ServiceSection) RuntimeSection {
	if service.Status != serviceStatusRunning {
		return RuntimeSection{Status: RuntimeStatusUnknown}
	}
	switch {
	case readiness.Status == "ready" && readiness.Database == "ok":
		return RuntimeSection{Status: RuntimeStatusReady}
	case readiness.Status == "not_ready" && readiness.Database == "error":
		return RuntimeSection{Status: RuntimeStatusDegraded, Note: RuntimeNoteDatabase}
	default:
		return RuntimeSection{Status: RuntimeStatusUnknown}
	}
}
