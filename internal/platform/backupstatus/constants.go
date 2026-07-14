package backupstatus

import "time"

const (
	// StatusFileName is the fixed backup status sidecar basename under state_dir.
	StatusFileName = "backup-last-run.json"
	statusTempName = "backup-last-run.json.tmp"

	SchemaVersion = 1

	MaxWriteSize = 1024
	MaxReadSize  = 4096

	// API status values.
	StatusNeverRun    = "never_run"
	StatusSucceeded   = "succeeded"
	StatusFailed      = "failed"
	StatusUnavailable = "unavailable"

	OutcomeSucceeded = "succeeded"
	OutcomeFailed    = "failed"

	FailureArtifact  = "artifact"
	FailureRestart   = "restart"
	FailureHealth    = "health"
	FailureReadiness = "readiness"
	FailureInternal  = "internal"

	statusFileMode   = 0o640
	vyntrioGroupName = "vyntrio"
)

const timestampSkew = 2 * time.Minute

var earliestValidTimestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
