package settings

// Snapshot holds validated system settings loaded at startup for read-only checks.
type Snapshot struct {
	Timezone string
	Hostname string
}

// NewSnapshot creates an immutable snapshot from validated system settings.
func NewSnapshot(sys SystemSettings) Snapshot {
	return Snapshot{
		Timezone: sys.Timezone,
		Hostname: sys.Hostname,
	}
}
