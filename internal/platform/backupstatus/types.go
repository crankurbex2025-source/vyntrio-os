package backupstatus

// Backup is the safe backup status section for GET /api/v1/overview.
type Backup struct {
	Status        string  `json:"status"`
	CompletedAt   *string `json:"completed_at,omitempty"`
	EverSucceeded *bool   `json:"ever_succeeded,omitempty"`
	Failure       *string `json:"failure,omitempty"`
}

// DiskRecord is the on-disk backup status sidecar schema version 1.
type DiskRecord struct {
	SchemaVersion int    `json:"schema_version"`
	LastOutcome   string `json:"last_outcome"`
	CompletedAt   string `json:"completed_at"`
	EverSucceeded bool   `json:"ever_succeeded"`
	FailureClass  string `json:"failure_class,omitempty"`
}

// NeverRun returns the API shape for a missing status record.
func NeverRun() Backup {
	ever := false
	return Backup{
		Status:        StatusNeverRun,
		EverSucceeded: &ever,
	}
}

// Unavailable returns the API shape when the status record cannot be trusted.
func Unavailable() Backup {
	return Backup{Status: StatusUnavailable}
}
