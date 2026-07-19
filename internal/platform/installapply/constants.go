package installapply

const (
	RecordDirName   = "install-apply"
	ApplyRecordName = "PARTITION_APPLY_RECORD.txt"
	SchemaVersion   = "vyntrio-install-partition-apply-v1"

	StatusApplied    = "applied"
	StatusRolledBack = "rolled_back"
	StatusFailed     = "failed"

	ScopeStatement = "existing_partition_payload_apply_only; six allowlisted payloads; not full OS install"
)
