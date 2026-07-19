package installpostflight

import (
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installapply"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installwrite"
)

// InputFromPreflight builds handover input after a preflight command.
func InputFromPreflight(
	targetDiskID, releaseManifest, envelopeRoot string,
	preflight installpreflight.Result,
	ok bool,
) Input {
	return Input{
		Command:             CommandPreflight,
		TargetDiskID:        targetDiskID,
		ReleaseManifestPath: releaseManifest,
		EnvelopeRoot:        envelopeRoot,
		TargetPreflight:     preflight.Target.Status,
		MediaEnvelope:       preflight.Media.EnvelopeStatus,
		ReleaseIntegrity:    preflight.Media.ReleaseIntegrity,
		ReleaseAuthenticity: preflight.Media.ReleaseAuth,
		PreflightOK:         ok,
		InstallWrite:        WriteNotRun,
		PayloadAllowlist:    len(installpreflight.RequiredPayloadRelativePaths()),
	}
}

// InputFromInstall builds handover input from an install outcome.
func InputFromInstall(
	releaseManifest, envelopeRoot string,
	outcome installwrite.Outcome,
	writeStatus string,
) Input {
	paths := append([]string(nil), outcome.PayloadPaths...)
	return Input{
		Command:             CommandInstall,
		TargetDiskID:        outcome.TargetDiskID,
		ReleaseManifestPath: releaseManifest,
		ReleaseVersion:      outcome.ReleaseVersion,
		EnvelopeRoot:        envelopeRoot,
		TargetPreflight:     outcome.Preflight.Target.Status,
		MediaEnvelope:       outcome.Preflight.Media.EnvelopeStatus,
		ReleaseIntegrity:    outcome.Preflight.Media.ReleaseIntegrity,
		ReleaseAuthenticity: outcome.Preflight.Media.ReleaseAuth,
		PreflightOK:         outcome.PreflightPassed,
		InstallWrite:        writeStatus,
		SandboxTargetPath:   outcome.TargetRoot,
		PayloadsCopied:      outcome.PayloadsCopied,
		PayloadAllowlist:    len(installpreflight.RequiredPayloadRelativePaths()),
		PayloadPaths:        paths,
		FailureStage:        outcome.FailureStage,
		FailureReason:       outcome.FailureReason,
		ApplyTarget:         outcome.ApplyTarget,
		HostBlockDeviceMutated: outcome.HostBlockDeviceMutated,
		TargetMutationRecord:   outcome.TargetMutationRecord,
		MountPoint:             outcome.MountPoint,
	}
}

// InputFromApply builds handover input from a partition apply outcome.
func InputFromApply(
	releaseManifest, envelopeRoot string,
	installOutcome installwrite.Outcome,
	applyOutcome installapply.Outcome,
	writeStatus string,
) Input {
	paths := append([]string(nil), installOutcome.PayloadPaths...)
	return Input{
		Command:                CommandApply,
		TargetDiskID:           installOutcome.TargetDiskID,
		ReleaseManifestPath:      releaseManifest,
		ReleaseVersion:         installOutcome.ReleaseVersion,
		EnvelopeRoot:             envelopeRoot,
		TargetPreflight:          installOutcome.Preflight.Target.Status,
		MediaEnvelope:            installOutcome.Preflight.Media.EnvelopeStatus,
		ReleaseIntegrity:         installOutcome.Preflight.Media.ReleaseIntegrity,
		ReleaseAuthenticity:      installOutcome.Preflight.Media.ReleaseAuth,
		PreflightOK:              installOutcome.PreflightPassed,
		InstallWrite:             writeStatus,
		SandboxTargetPath:        applyOutcome.MountPoint,
		PayloadsCopied:           installOutcome.PayloadsCopied,
		PayloadAllowlist:         len(installpreflight.RequiredPayloadRelativePaths()),
		PayloadPaths:             paths,
		FailureStage:             installOutcome.FailureStage,
		FailureReason:            installOutcome.FailureReason,
		ApplyTarget:              true,
		HostBlockDeviceMutated:   installOutcome.HostBlockDeviceMutated,
		TargetMutationRecord:     applyOutcome.MutationRecordPath,
		PartitionApplyRecord:     applyOutcome.ApplyRecordPath,
		MountPoint:               applyOutcome.MountPoint,
		PartitionDevicePath:      applyOutcome.PartitionDevicePath,
	}
}

// FailureReasonFromError maps install errors to operator-facing reasons.
func FailureReasonFromError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, string(installwrite.ErrForceRequired.Error())):
		return installwrite.StageForce
	case strings.Contains(msg, string(installwrite.ErrArtifactSourceRequired.Error())):
		return installwrite.StageArtifactSource
	case strings.Contains(msg, string(installwrite.ErrUnsafeTargetRoot.Error())):
		return installwrite.StageTargetRoot
	case strings.Contains(msg, string(installwrite.ErrPreflightFailed.Error())):
		return "preflight_failed"
	case strings.Contains(msg, string(installwrite.ErrPostVerifyFailed.Error())):
		return "post_verify_failed"
	default:
		return strings.TrimSpace(msg)
	}
}
