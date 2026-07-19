package installapply

import (
	"context"
	"fmt"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installtarget"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installwrite"
)

const (
	StageForce          = "force"
	StageTargetID       = "target_disk_id"
	StageArtifactSource = "artifact_source"
	StagePreflight      = "preflight"
	StageTargetMount    = "target_mount"
	StageCopyPlan       = "copy_plan"
	StageStateDirs      = "state_directories"
	StagePayloadCopy    = "payload_copy"
	StagePostVerify     = "post_verify"
	StageInstallRecord  = "install_record"
	StageApplyRecord    = "apply_record"
	StageTargetUnmount  = "target_unmount"
)

// Runner performs guarded existing-partition payload application.
type Runner struct {
	Checker     installpreflight.Checker
	Applier     installtarget.Applier
	CopyPayload func(entry installwrite.CopyEntry, targetRoot string) error
}

// NewRunner creates a runner with platform defaults.
func NewRunner(stateDir string) Runner {
	if strings.TrimSpace(stateDir) == "" {
		stateDir = installpreflight.DefaultStateDir
	}
	return Runner{
		Checker: installpreflight.NewChecker(stateDir),
		Applier: installtarget.NewApplier(stateDir),
	}
}

// Run executes preflight, mount, allowlisted payload copy, verify, records, and unmount.
func (r Runner) Run(ctx context.Context, req Request) (Outcome, error) {
	_ = ctx
	outcome := Outcome{}

	if !req.Force {
		outcome.FailureStage = StageForce
		outcome.FailureReason = ErrForceRequired.Error()
		return outcome, ErrForceRequired
	}

	diskID := strings.TrimSpace(req.TargetDiskID)
	if diskID == "" {
		outcome.FailureStage = StageTargetID
		outcome.FailureReason = "target disk id required"
		return outcome, fmt.Errorf("%w: target disk id required", ErrApplyFailed)
	}
	outcome.TargetDiskID = diskID

	if strings.TrimSpace(req.EnvelopeRoot) == "" && strings.TrimSpace(req.ReleaseManifest) == "" {
		outcome.FailureStage = StageArtifactSource
		outcome.FailureReason = ErrArtifactRequired.Error()
		return outcome, ErrArtifactRequired
	}

	stateDir := strings.TrimSpace(req.StateDir)
	if stateDir == "" {
		stateDir = installpreflight.DefaultStateDir
	}

	preflightResult, err := r.Checker.Run(context.Background(),
		installpreflight.TargetRequest{DiskID: diskID, StateDir: stateDir},
		installpreflight.MediaRequest{
			EnvelopeRoot:        strings.TrimSpace(req.EnvelopeRoot),
			ReleaseManifestPath: strings.TrimSpace(req.ReleaseManifest),
			ArtifactBaseDir:     strings.TrimSpace(req.ArtifactBaseDir),
		},
	)
	outcome.Preflight = preflightResult
	if err != nil {
		outcome.FailureStage = StagePreflight
		outcome.FailureReason = err.Error()
		return outcome, fmt.Errorf("%w: %v", ErrPreflightFailed, err)
	}
	outcome.PreflightPassed = true

	mountSession, applyOutcome, err := r.Applier.Prepare(installtarget.ApplyRequest{
		DiskID:    diskID,
		StateDir:  stateDir,
		MountRoot: strings.TrimSpace(req.MountRoot),
		Force:     true,
	})
	outcome.MountPoint = applyOutcome.MountPoint
	outcome.TargetFSType = applyOutcome.FSType
	outcome.PartitionDevicePath = applyOutcome.Candidate.DevicePath
	if err != nil {
		outcome.FailureStage = applyOutcome.FailureStage
		if outcome.FailureStage == "" {
			outcome.FailureStage = StageTargetMount
		}
		outcome.FailureReason = applyOutcome.FailureReason
		if outcome.FailureReason == "" {
			outcome.FailureReason = err.Error()
		}
		outcome.MutationRecordPath, _ = installtarget.WriteMutationRecord(stateDir, applyOutcome, installtarget.StatusFailed, 0, nil)
		outcome.ApplyRecordPath, _ = WriteApplyRecord(stateDir, outcome, StatusFailed)
		return outcome, err
	}

	targetRoot := mountSession.MountPoint
	plan, releaseVersion, err := installwrite.BuildCopyPlan(req.EnvelopeRoot, req.ReleaseManifest, req.ArtifactBaseDir)
	if err != nil {
		outcome.FailureStage = StageCopyPlan
		outcome.FailureReason = err.Error()
		r.rollbackAfterFailure(mountSession, &applyOutcome, targetRoot, nil, stateDir, &outcome)
		return outcome, err
	}
	outcome.ReleaseVersion = releaseVersion

	if err := installwrite.CreateStateDirectories(targetRoot); err != nil {
		outcome.FailureStage = StageStateDirs
		outcome.FailureReason = err.Error()
		r.rollbackAfterFailure(mountSession, &applyOutcome, targetRoot, nil, stateDir, &outcome)
		return outcome, fmt.Errorf("%w: %v", ErrApplyFailed, err)
	}

	for _, entry := range plan {
		copyFn := installwrite.CopyPayloadFile
		if r.CopyPayload != nil {
			copyFn = r.CopyPayload
		}
		if err := copyFn(entry, targetRoot); err != nil {
			outcome.FailureStage = StagePayloadCopy
			outcome.FailureReason = err.Error()
			outcome.PayloadsCopied = countCopied(plan, entry.TargetRel)
			outcome.PayloadPaths = copiedPaths(plan, entry.TargetRel)
			r.rollbackAfterFailure(mountSession, &applyOutcome, targetRoot, outcome.PayloadPaths, stateDir, &outcome)
			return outcome, fmt.Errorf("%w: %v", ErrApplyFailed, err)
		}
		outcome.PayloadPaths = append(outcome.PayloadPaths, entry.TargetRel)
	}
	outcome.PayloadsCopied = len(plan)

	if err := installwrite.PostVerifyTarget(targetRoot, plan); err != nil {
		outcome.FailureStage = StagePostVerify
		outcome.FailureReason = err.Error()
		r.rollbackAfterFailure(mountSession, &applyOutcome, targetRoot, outcome.PayloadPaths, stateDir, &outcome)
		return outcome, fmt.Errorf("%w: %v", ErrApplyFailed, err)
	}

	if err := installwrite.WriteInstallRecordFile(targetRoot, diskID, releaseVersion, plan, true); err != nil {
		outcome.FailureStage = StageInstallRecord
		outcome.FailureReason = err.Error()
		r.rollbackAfterFailure(mountSession, &applyOutcome, targetRoot, append(outcome.PayloadPaths, installwrite.InstallRecordName), stateDir, &outcome)
		return outcome, fmt.Errorf("%w: %v", ErrApplyFailed, err)
	}

	if err := mountSession.Unmount(); err != nil {
		outcome.FailureStage = StageTargetUnmount
		outcome.FailureReason = err.Error()
		outcome.MutationRecordPath, _ = installtarget.WriteMutationRecord(stateDir, applyOutcome, installtarget.StatusFailed, outcome.PayloadsCopied, outcome.PayloadPaths)
		outcome.ApplyRecordPath, _ = WriteApplyRecord(stateDir, outcome, StatusFailed)
		return outcome, fmt.Errorf("%w: unmount: %v", ErrApplyFailed, err)
	}
	mountSession.MountPoint = ""

	outcome.HostBlockDeviceMutated = true
	outcome.MutationRecordPath, err = installtarget.WriteMutationRecord(stateDir, applyOutcome, installtarget.StatusApplied, outcome.PayloadsCopied, outcome.PayloadPaths)
	if err != nil {
		outcome.FailureStage = StageApplyRecord
		outcome.FailureReason = err.Error()
		return outcome, err
	}
	outcome.ApplyRecordPath, err = WriteApplyRecord(stateDir, outcome, StatusApplied)
	if err != nil {
		outcome.FailureStage = StageApplyRecord
		outcome.FailureReason = err.Error()
		return outcome, err
	}
	return outcome, nil
}

func (r Runner) rollbackAfterFailure(
	session installtarget.MountSession,
	applyOutcome *installtarget.ApplyOutcome,
	mountPoint string,
	payloadPaths []string,
	stateDir string,
	outcome *Outcome,
) {
	if len(payloadPaths) > 0 {
		installtarget.RollbackCopiedFiles(mountPoint, payloadPaths)
	}
	_ = installtarget.RollbackUnmount(session, applyOutcome)
	outcome.MutationRecordPath, _ = installtarget.WriteMutationRecord(stateDir, *applyOutcome, installtarget.StatusRolledBack, outcome.PayloadsCopied, outcome.PayloadPaths)
	outcome.ApplyRecordPath, _ = WriteApplyRecord(stateDir, *outcome, StatusRolledBack)
}

func countCopied(plan []installwrite.CopyEntry, failedTarget string) int {
	count := 0
	for _, entry := range plan {
		if entry.TargetRel == failedTarget {
			break
		}
		count++
	}
	return count
}

func copiedPaths(plan []installwrite.CopyEntry, failedTarget string) []string {
	var paths []string
	for _, entry := range plan {
		if entry.TargetRel == failedTarget {
			break
		}
		paths = append(paths, entry.TargetRel)
	}
	return paths
}

// ToInstallOutcome maps apply outcome into installwrite outcome for shared handover.
func ToInstallOutcome(req Request, outcome Outcome) installwrite.Outcome {
	return installwrite.Outcome{
		TargetDiskID:           outcome.TargetDiskID,
		TargetRoot:             outcome.MountPoint,
		PayloadsCopied:         outcome.PayloadsCopied,
		PayloadPaths:           append([]string(nil), outcome.PayloadPaths...),
		ReleaseVersion:         outcome.ReleaseVersion,
		Preflight:              outcome.Preflight,
		PreflightPassed:        outcome.PreflightPassed,
		ApplyTarget:            true,
		HostBlockDeviceMutated: outcome.HostBlockDeviceMutated,
		TargetMutationRecord:   outcome.MutationRecordPath,
		MountPoint:             outcome.MountPoint,
		TargetFSType:           outcome.TargetFSType,
		FailureStage:           outcome.FailureStage,
		FailureReason:          outcome.FailureReason,
	}
}
