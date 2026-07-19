package installwrite

import (
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
)

const (
	StageForce           = "force"
	StageTargetID        = "target_disk_id"
	StageArtifactSource  = "artifact_source"
	StageTargetRoot      = "target_root"
	StagePreflight       = "preflight"
	StageCopyPlan        = "copy_plan"
	StageStateDirs       = "state_directories"
	StagePayloadCopy     = "payload_copy"
	StagePostVerify      = "post_verify"
	StageInstallRecord   = "install_record"
	StageTargetMount     = "target_mount"
	StageTargetRollback  = "target_rollback"
	StageTargetMutation  = "target_mutation_record"
)

// Outcome captures install execution facts for postflight handover.
type Outcome struct {
	TargetDiskID            string
	TargetRoot              string
	PayloadsCopied          int
	PayloadPaths            []string
	ReleaseVersion          string
	Preflight               installpreflight.Result
	PreflightPassed         bool
	ApplyTarget             bool
	HostBlockDeviceMutated  bool
	TargetMutationRecord    string
	MountPoint              string
	TargetFSType            string
	FailureStage            string
	FailureReason           string
}
