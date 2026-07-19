package installpostflight_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpostflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installwrite"
)

func TestBuildPreflightOnlyHandover(t *testing.T) {
	input := installpostflight.InputFromPreflight(
		"disk-abc",
		"/tmp/manifest.json",
		"/tmp/envelope",
		installpreflight.Result{
			Target: installpreflight.TargetResult{Status: installpreflight.TargetEligible, DiskID: "disk-abc"},
			Media: installpreflight.MediaResult{
				EnvelopeStatus:   installpreflight.MediaOK,
				ReleaseIntegrity: installpreflight.ReleaseOK,
			},
		},
		true,
	)
	h := installpostflight.Build(input)
	if h.OverallStatus != installpostflight.OverallPreflightOnly {
		t.Fatalf("overall = %q", h.OverallStatus)
	}
	if h.InstallWrite != installpostflight.WriteNotRun {
		t.Fatalf("install_write = %q", h.InstallWrite)
	}
	if h.MutationScope != installpostflight.MutationScopeStatement {
		t.Fatal("expected mutation scope statement")
	}
}

func TestBuildInstallSuccessHandover(t *testing.T) {
	input := installpostflight.InputFromInstall(
		"/tmp/manifest.json",
		"/tmp/envelope",
		installwrite.Outcome{
			TargetDiskID:    "disk-abc",
			TargetRoot:      "/sandbox/disk-abc",
			PayloadsCopied:  6,
			PayloadPaths:    []string{"usr/bin/vyntrio-api"},
			ReleaseVersion:  "0.2.0-dev",
			PreflightPassed: true,
			Preflight: installpreflight.Result{
				Target: installpreflight.TargetResult{Status: installpreflight.TargetEligible},
				Media: installpreflight.MediaResult{
					EnvelopeStatus:   installpreflight.MediaOK,
					ReleaseIntegrity: installpreflight.ReleaseOK,
				},
			},
		},
		installpostflight.WriteSucceeded,
	)
	h := installpostflight.Build(input)
	if h.OverallStatus != installpostflight.OverallSucceeded {
		t.Fatalf("overall = %q", h.OverallStatus)
	}
	if !h.AllowlistComplete {
		t.Fatal("expected allowlist complete")
	}
}

func TestBuildInstallFailureHandover(t *testing.T) {
	input := installpostflight.InputFromInstall(
		"",
		"/tmp/envelope",
		installwrite.Outcome{
			TargetDiskID:    "disk-abc",
			TargetRoot:      "/sandbox/disk-abc",
			PreflightPassed: true,
			FailureStage:    installwrite.StagePostVerify,
			FailureReason:   "sha256 mismatch",
		},
		installpostflight.WriteFailed,
	)
	h := installpostflight.Build(input)
	if h.OverallStatus != installpostflight.OverallFailed {
		t.Fatalf("overall = %q", h.OverallStatus)
	}
	if h.FailureStage != installwrite.StagePostVerify {
		t.Fatalf("stage = %q", h.FailureStage)
	}
}

func TestWriteAndReadRecordRoundtrip(t *testing.T) {
	dir := t.TempDir()
	h := installpostflight.Handover{
		SchemaVersion:     installpostflight.SchemaVersion,
		GeneratedAt:       time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
		Command:           installpostflight.CommandInstall,
		OverallStatus:     installpostflight.OverallSucceeded,
		TargetDiskID:      "disk-abc",
		PreflightStatus:   installpostflight.PreflightOK,
		InstallWrite:      installpostflight.WriteSucceeded,
		SandboxTargetPath: dir,
		PayloadsCopied:    6,
		PayloadAllowlist:  6,
		AllowlistComplete: true,
		MutationScope:     installpostflight.MutationScopeStatement,
		DeferredItems:     []string{"bootstrap_handoff"},
		NextSteps:         []string{"Review records."},
		PayloadPaths:      []string{"usr/bin/vyntrio-api"},
	}
	if _, err := installpostflight.WriteRecord(dir, h); err != nil {
		t.Fatalf("WriteRecord() error: %v", err)
	}
	got, err := installpostflight.ReadRecord(dir)
	if err != nil {
		t.Fatalf("ReadRecord() error: %v", err)
	}
	if got.TargetDiskID != "disk-abc" {
		t.Fatalf("disk_id = %q", got.TargetDiskID)
	}
	if got.PayloadsCopied != 6 {
		t.Fatalf("payloads = %d", got.PayloadsCopied)
	}
	if _, err := os.Stat(filepath.Join(dir, installpostflight.HandoverRecordName)); err != nil {
		t.Fatalf("missing handover record: %v", err)
	}
}

func TestSummaryLineIncludesMutationScope(t *testing.T) {
	h := installpostflight.Build(installpostflight.Input{
		Command:       installpostflight.CommandInstall,
		TargetDiskID:  "disk-abc",
		PreflightOK:   true,
		InstallWrite:  installpostflight.WriteSucceeded,
		PayloadsCopied: 6,
		PayloadAllowlist: 6,
	})
	line := installpostflight.SummaryLine(h)
	if !strings.Contains(line, "mutation_scope=") {
		t.Fatalf("summary = %q", line)
	}
}
