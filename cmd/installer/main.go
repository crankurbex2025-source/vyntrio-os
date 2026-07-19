package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installapply"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpolicy"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpostflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installtarget"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installwrite"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "help", "-h", "--help":
		usage()
		return 0
	case "workflow":
		return runWorkflow()
	case "preflight":
		return runPreflight(args[1:])
	case "install":
		return runInstall(args[1:])
	case "apply":
		return runApply(args[1:])
	case "postflight":
		return runPostflight(args[1:])
	default:
		usage()
		return 2
	}
}

func runWorkflow() int {
	_, _ = fmt.Fprintln(os.Stdout, installpolicy.WorkflowText())
	return 0
}

func runPreflight(args []string) int {
	fs := flag.NewFlagSet("preflight", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	targetDiskID := fs.String("target-disk-id", "", "opaque install target disk ID (required)")
	stateDir := fs.String("state-dir", installpreflight.DefaultStateDir, "state directory for discovery context")
	minSize := fs.String("min-size-bytes", "", "minimum target size in bytes (default: 8GiB)")
	envelopeRoot := fs.String("envelope-root", "", "install-media envelope root for media checks")
	releaseManifest := fs.String("release-manifest", "", "release manifest path for artifact integrity checks")
	artifactBaseDir := fs.String("artifact-base-dir", "", "base directory for release manifest artifacts")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		usage()
		return 2
	}

	targetReq := installpreflight.TargetRequest{
		DiskID:   strings.TrimSpace(*targetDiskID),
		StateDir: strings.TrimSpace(*stateDir),
	}
	if strings.TrimSpace(*minSize) != "" {
		parsed, err := strconv.ParseUint(strings.TrimSpace(*minSize), 10, 64)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "installer preflight failed: reason=invalid_min_size_bytes\n")
			return 1
		}
		targetReq.MinSizeBytes = parsed
	}

	mediaReq := installpreflight.MediaRequest{
		EnvelopeRoot:        strings.TrimSpace(*envelopeRoot),
		ReleaseManifestPath: strings.TrimSpace(*releaseManifest),
		ArtifactBaseDir:     strings.TrimSpace(*artifactBaseDir),
	}

	checker := installpreflight.NewChecker(targetReq.StateDir)
	result, err := checker.Run(nil, targetReq, mediaReq)
	if err != nil {
		printTargetFailure(result.Target)
		for _, failure := range result.Media.Failures {
			_, _ = fmt.Fprintf(os.Stderr, "installer preflight failed: scope=%s reason=%s detail=%s\n",
				failure.Scope, failure.Reason, failure.Detail)
		}
		handover := installpostflight.Build(installpostflight.InputFromPreflight(
			targetReq.DiskID,
			mediaReq.ReleaseManifestPath,
			mediaReq.EnvelopeRoot,
			result,
			false,
		))
		printPostflight(handover)
		return 1
	}

	handover := installpostflight.Build(installpostflight.InputFromPreflight(
		targetReq.DiskID,
		mediaReq.ReleaseManifestPath,
		mediaReq.EnvelopeRoot,
		result,
		true,
	))
	printPostflight(handover)
	return 0
}

func runInstall(args []string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	targetDiskID := fs.String("target-disk-id", "", "opaque install target disk ID (required)")
	stateDir := fs.String("state-dir", installpreflight.DefaultStateDir, "state directory for discovery context")
	envelopeRoot := fs.String("envelope-root", "", "install-media envelope root")
	releaseManifest := fs.String("release-manifest", "", "release manifest path")
	artifactBaseDir := fs.String("artifact-base-dir", "", "base directory for release manifest artifacts")
	sandboxRoot := fs.String("sandbox-root", installwrite.DefaultSandboxRoot, "install sandbox root")
	targetRoot := fs.String("target-root", "", "optional target root under sandbox (default: <sandbox>/<disk-id>)")
	force := fs.Bool("force", false, "required to authorize payload writes")
	applyTarget := fs.Bool("apply-target", false, "mount eligible target and write payloads to real block device (requires --force)")
	mountRoot := fs.String("mount-root", installtarget.DefaultMountRoot, "temporary mount root for --apply-target")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		usage()
		return 2
	}

	if *applyTarget {
		return executePartitionApply(installapply.Request{
			TargetDiskID:    strings.TrimSpace(*targetDiskID),
			StateDir:        strings.TrimSpace(*stateDir),
			EnvelopeRoot:    strings.TrimSpace(*envelopeRoot),
			ReleaseManifest: strings.TrimSpace(*releaseManifest),
			ArtifactBaseDir: strings.TrimSpace(*artifactBaseDir),
			MountRoot:       strings.TrimSpace(*mountRoot),
			Force:           *force,
		}, installpostflight.CommandInstall)
	}

	req := installwrite.Request{
		TargetDiskID:    strings.TrimSpace(*targetDiskID),
		TargetRoot:      strings.TrimSpace(*targetRoot),
		SandboxRoot:     strings.TrimSpace(*sandboxRoot),
		StateDir:        strings.TrimSpace(*stateDir),
		EnvelopeRoot:    strings.TrimSpace(*envelopeRoot),
		ReleaseManifest: strings.TrimSpace(*releaseManifest),
		ArtifactBaseDir: strings.TrimSpace(*artifactBaseDir),
		Force:           *force,
	}

	installer := installwrite.NewInstaller(req.StateDir)
	outcome, err := installer.Install(nil, req)
	writeStatus := installpostflight.WriteFailed
	if err == nil {
		writeStatus = installpostflight.WriteSucceeded
	} else if outcome.PayloadsCopied > 0 && outcome.PayloadsCopied < len(installpreflight.RequiredPayloadRelativePaths()) {
		writeStatus = installpostflight.WritePartial
	}

	handover := installpostflight.Build(installpostflight.InputFromInstall(
		req.ReleaseManifest,
		req.EnvelopeRoot,
		outcome,
		writeStatus,
	))

	handoverRoot := outcome.TargetRoot
	if handoverRoot != "" {
		if _, writeErr := installpostflight.WriteRecord(handoverRoot, handover); writeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "installer install warning: handover record not written: %v\n", writeErr)
		}
	}

	if err != nil {
		printInstallFailure(err)
		printPostflight(handover)
		return 1
	}

	printPostflight(handover)
	return 0
}

func runApply(args []string) int {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	targetDiskID := fs.String("target-disk-id", "", "opaque install target disk ID (required)")
	stateDir := fs.String("state-dir", installpreflight.DefaultStateDir, "state directory for discovery context")
	envelopeRoot := fs.String("envelope-root", "", "install-media envelope root")
	releaseManifest := fs.String("release-manifest", "", "release manifest path")
	artifactBaseDir := fs.String("artifact-base-dir", "", "base directory for release manifest artifacts")
	mountRoot := fs.String("mount-root", installtarget.DefaultMountRoot, "temporary mount root")
	force := fs.Bool("force", false, "required to authorize partition apply")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		usage()
		return 2
	}

	req := installapply.Request{
		TargetDiskID:    strings.TrimSpace(*targetDiskID),
		StateDir:        strings.TrimSpace(*stateDir),
		EnvelopeRoot:    strings.TrimSpace(*envelopeRoot),
		ReleaseManifest: strings.TrimSpace(*releaseManifest),
		ArtifactBaseDir: strings.TrimSpace(*artifactBaseDir),
		MountRoot:       strings.TrimSpace(*mountRoot),
		Force:           *force,
	}

	return executePartitionApply(req, installpostflight.CommandApply)
}

func executePartitionApply(req installapply.Request, command string) int {
	runner := installapply.NewRunner(req.StateDir)
	outcome, err := runner.Run(nil, req)
	writeStatus := installpostflight.WriteFailed
	if err == nil {
		writeStatus = installpostflight.WriteSucceeded
	} else if outcome.PayloadsCopied > 0 && outcome.PayloadsCopied < len(installpreflight.RequiredPayloadRelativePaths()) {
		writeStatus = installpostflight.WritePartial
	}

	installOutcome := installapply.ToInstallOutcome(req, outcome)
	handover := installpostflight.Build(installpostflight.InputFromApply(
		req.ReleaseManifest,
		req.EnvelopeRoot,
		installOutcome,
		outcome,
		writeStatus,
	))
	if command == installpostflight.CommandInstall {
		handover.Command = installpostflight.CommandInstall
	}

	handoverRoot := ""
	if outcome.ApplyRecordPath != "" {
		handoverRoot = filepath.Dir(outcome.ApplyRecordPath)
	}
	if handoverRoot != "" {
		if _, writeErr := installpostflight.WriteRecord(handoverRoot, handover); writeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "installer apply warning: handover record not written: %v\n", writeErr)
		}
	}

	if err != nil {
		if command == installpostflight.CommandApply {
			printApplyFailure(err)
		} else {
			printInstallFailure(err)
		}
		printPostflight(handover)
		return 1
	}

	printPostflight(handover)
	return 0
}

func runPostflight(args []string) int {
	fs := flag.NewFlagSet("postflight", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	targetRoot := fs.String("target-root", "", "directory containing HANDOVER_RECORD.txt (sandbox or install-apply record dir)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		usage()
		return 2
	}
	if strings.TrimSpace(*targetRoot) == "" {
		_, _ = fmt.Fprintln(os.Stderr, "installer postflight failed: reason=target_root_required")
		return 1
	}

	handover, err := installpostflight.ReadRecord(strings.TrimSpace(*targetRoot))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "installer postflight failed: reason=handover_record_missing detail=%v\n", err)
		return 1
	}
	printPostflight(handover)
	if handover.OverallStatus == installpostflight.OverallFailed {
		return 1
	}
	return 0
}

func printPostflight(h installpostflight.Handover) {
	_, _ = fmt.Fprintln(os.Stdout, installpostflight.SummaryLine(h))
	for _, line := range installpostflight.NoteLines(h) {
		_, _ = fmt.Fprintln(os.Stderr, line)
	}
}

func printApplyFailure(err error) {
	switch {
	case errors.Is(err, installapply.ErrForceRequired):
		_, _ = fmt.Fprintln(os.Stderr, "installer apply failed: reason=force_required")
		printPolicyHint()
	case errors.Is(err, installapply.ErrArtifactRequired):
		_, _ = fmt.Fprintln(os.Stderr, "installer apply failed: reason=artifact_source_required")
		printPolicyHint()
	case errors.Is(err, installapply.ErrPreflightFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer apply failed: reason=preflight_failed detail=%v\n", err)
	case errors.Is(err, installtarget.ErrTargetMounted):
		_, _ = fmt.Fprintln(os.Stderr, "installer apply failed: reason=target_mounted")
	case errors.Is(err, installtarget.ErrTargetNotEligible):
		_, _ = fmt.Fprintln(os.Stderr, "installer apply failed: reason=target_not_eligible")
	case errors.Is(err, installtarget.ErrAmbiguousTargetLayout):
		_, _ = fmt.Fprintln(os.Stderr, "installer apply failed: reason=ambiguous_target_layout")
	case errors.Is(err, installtarget.ErrUnsupportedTargetState):
		_, _ = fmt.Fprintln(os.Stderr, "installer apply failed: reason=unsupported_target_state")
	case errors.Is(err, installtarget.ErrMountFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer apply failed: reason=mount_failed detail=%v\n", err)
	case errors.Is(err, installapply.ErrApplyFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer apply failed: detail=%v\n", err)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "installer apply failed: %v\n", err)
	}
}

func printInstallFailure(err error) {
	switch {
	case errors.Is(err, installwrite.ErrForceRequired):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=force_required")
		printPolicyHint()
	case errors.Is(err, installwrite.ErrArtifactSourceRequired):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=artifact_source_required")
		printPolicyHint()
	case errors.Is(err, installwrite.ErrUnsafeTargetRoot):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=unsafe_target_root")
	case errors.Is(err, installwrite.ErrPreflightFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer install failed: reason=preflight_failed detail=%v\n", err)
	case errors.Is(err, installwrite.ErrPostVerifyFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer install failed: reason=post_verify_failed detail=%v\n", err)
	case errors.Is(err, installtarget.ErrForceRequired):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=force_required")
	case errors.Is(err, installtarget.ErrTargetMounted):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=target_mounted")
	case errors.Is(err, installtarget.ErrTargetNotEligible):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=target_not_eligible")
	case errors.Is(err, installtarget.ErrAmbiguousTargetLayout):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=ambiguous_target_layout")
	case errors.Is(err, installtarget.ErrUnsupportedTargetState):
		_, _ = fmt.Fprintln(os.Stderr, "installer install failed: reason=unsupported_target_state")
	case errors.Is(err, installtarget.ErrMountFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer install failed: reason=mount_failed detail=%v\n", err)
	case errors.Is(err, installtarget.ErrRollbackFailed):
		_, _ = fmt.Fprintf(os.Stderr, "installer install failed: reason=rollback_failed detail=%v\n", err)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "installer install failed: %v\n", err)
	}
}

func printPolicyHint() {
	_, _ = fmt.Fprintln(os.Stderr, "installer hint: run 'vyntrio-installer workflow' for USB-first product path vs Block 10 internal infrastructure boundaries")
}

func printTargetFailure(target installpreflight.TargetResult) {
	if target.DiskID == "" && len(target.Reasons) == 1 && target.Reasons[0] == installpreflight.ReasonTargetSelectionRequired {
		_, _ = fmt.Fprintf(os.Stderr, "installer preflight failed: scope=target reason=%s\n", installpreflight.ReasonTargetSelectionRequired)
		printPolicyHint()
		return
	}
	reason := "unknown"
	if len(target.Reasons) > 0 {
		reason = target.Reasons[0]
	}
	_, _ = fmt.Fprintf(os.Stderr, "installer preflight failed: scope=target status=%s reason=%s disk_id=%s\n",
		target.Status, reason, target.DiskID)
	if len(target.Reasons) > 1 {
		for _, extra := range target.Reasons[1:] {
			_, _ = fmt.Fprintf(os.Stderr, "installer preflight failed: scope=target reason=%s disk_id=%s\n", extra, target.DiskID)
		}
	}
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, installpolicy.UsageText())
}
