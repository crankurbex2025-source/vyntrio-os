package installtarget

const (
	DefaultMountRoot = "/run/vyntrio-install/mnt"
	RecordDirName    = "install-target"
	MutationRecordName = "TARGET_MUTATION_RECORD.txt"
	SchemaVersion    = "vyntrio-install-target-mutation-v1"

	StatusPrepared = "prepared"
	StatusApplied  = "applied"
	StatusRolledBack = "rolled_back"
	StatusFailed   = "failed"

	ReasonUnsupportedTargetState = "unsupported_target_state"
	ReasonAmbiguousTargetLayout  = "ambiguous_target_layout"
	ReasonTargetMounted          = "target_mounted"
	ReasonTargetNotEligible      = "target_not_eligible"
	ReasonMountFailed            = "mount_failed"
	ReasonRollbackSucceeded      = "rollback_succeeded"
	ReasonRollbackFailed         = "rollback_failed"
)

var supportedFSTypes = map[string]struct{}{
	"ext4":  {},
	"xfs":   {},
	"btrfs": {},
}
