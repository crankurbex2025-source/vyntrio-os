package installwrite

import (
	"context"
	"fmt"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
)

// Installer performs bounded sandbox install payload writes after preflight gates.
type Installer struct {
	Checker installpreflight.Checker
}

// NewInstaller creates an installer with the default preflight checker.
func NewInstaller(stateDir string) Installer {
	return Installer{
		Checker: installpreflight.NewChecker(stateDir),
	}
}

// Install runs preflight, copies supported payloads, and post-verifies the sandbox target tree.
func (i Installer) Install(ctx context.Context, req Request) (Outcome, error) {
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
		return outcome, fmt.Errorf("%w: target disk id required", ErrInstallFailed)
	}
	outcome.TargetDiskID = diskID
	if strings.TrimSpace(req.EnvelopeRoot) == "" && strings.TrimSpace(req.ReleaseManifest) == "" {
		outcome.FailureStage = StageArtifactSource
		outcome.FailureReason = ErrArtifactSourceRequired.Error()
		return outcome, ErrArtifactSourceRequired
	}

	stateDir := strings.TrimSpace(req.StateDir)
	if stateDir == "" {
		stateDir = installpreflight.DefaultStateDir
	}

	preflightResult, err := i.Checker.Run(context.Background(),
		installpreflight.TargetRequest{
			DiskID:   diskID,
			StateDir: stateDir,
		},
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

	sandboxRoot := strings.TrimSpace(req.SandboxRoot)
	if sandboxRoot == "" {
		sandboxRoot = DefaultSandboxRoot
	}
	targetRoot, err := resolveTargetRoot(sandboxRoot, diskID, req.TargetRoot)
	if err != nil {
		outcome.FailureStage = StageTargetRoot
		outcome.FailureReason = err.Error()
		return outcome, err
	}
	outcome.TargetRoot = targetRoot

	plan, releaseVersion, err := buildCopyPlan(req.EnvelopeRoot, req.ReleaseManifest, req.ArtifactBaseDir)
	if err != nil {
		outcome.FailureStage = StageCopyPlan
		outcome.FailureReason = err.Error()
		return outcome, err
	}
	outcome.ReleaseVersion = releaseVersion

	if err := createStateDirectories(targetRoot); err != nil {
		outcome.FailureStage = StageStateDirs
		outcome.FailureReason = err.Error()
		return outcome, err
	}

	for _, entry := range plan {
		if err := copyPayload(entry, targetRoot); err != nil {
			outcome.FailureStage = StagePayloadCopy
			outcome.FailureReason = err.Error()
			outcome.PayloadsCopied = countCopied(plan, entry.TargetRel)
			outcome.PayloadPaths = copiedPaths(plan, entry.TargetRel)
			return outcome, err
		}
		outcome.PayloadPaths = append(outcome.PayloadPaths, entry.TargetRel)
	}
	outcome.PayloadsCopied = len(plan)

	if err := postVerify(targetRoot, plan); err != nil {
		outcome.FailureStage = StagePostVerify
		outcome.FailureReason = err.Error()
		return outcome, err
	}

	if err := writeInstallRecord(targetRoot, diskID, releaseVersion, plan, false); err != nil {
		outcome.FailureStage = StageInstallRecord
		outcome.FailureReason = err.Error()
		return outcome, err
	}

	return outcome, nil
}

func countCopied(plan []CopyEntry, failedTarget string) int {
	count := 0
	for _, entry := range plan {
		if entry.TargetRel == failedTarget {
			break
		}
		count++
	}
	return count
}

func copiedPaths(plan []CopyEntry, failedTarget string) []string {
	var paths []string
	for _, entry := range plan {
		if entry.TargetRel == failedTarget {
			break
		}
		paths = append(paths, entry.TargetRel)
	}
	return paths
}
