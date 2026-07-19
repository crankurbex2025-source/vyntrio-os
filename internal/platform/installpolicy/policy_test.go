package installpolicy_test

import (
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpolicy"
)

func TestUsageTextContract(t *testing.T) {
	text := installpolicy.UsageText()
	for _, marker := range installpolicy.RequiredUsageMarkers() {
		if !strings.Contains(text, marker) {
			t.Fatalf("usage missing marker %q", marker)
		}
	}
}

func TestWorkflowTextContract(t *testing.T) {
	text := installpolicy.WorkflowText()
	for _, marker := range installpolicy.RequiredWorkflowMarkers() {
		if !strings.Contains(text, marker) {
			t.Fatalf("workflow missing marker %q", marker)
		}
	}
}

func TestVerifyArtifactUsageReferencesUSBFirst(t *testing.T) {
	text := installpolicy.VerifyArtifactUsageText()
	for _, marker := range []string{
		"USB-first",
		"NOT the primary install journey",
		"Does NOT write disks",
		"preflight",
		"install",
		"apply",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("verify-artifact usage missing %q", marker)
		}
	}
}

func TestInstallCommandSummaryIsDevLabOnly(t *testing.T) {
	for _, cmd := range installpolicy.Commands() {
		if cmd.Name != "install" {
			continue
		}
		if !strings.Contains(cmd.Summary, "Dev/lab only") {
			t.Fatalf("install summary = %q", cmd.Summary)
		}
		return
	}
	t.Fatal("install command missing")
}

func TestCommandsCoverStagedPipeline(t *testing.T) {
	commands := installpolicy.Commands()
	if len(commands) != 5 {
		t.Fatalf("commands = %d", len(commands))
	}
	names := make(map[string]struct{})
	for _, cmd := range commands {
		names[cmd.Name] = struct{}{}
	}
	for _, want := range []string{"verify-artifact", "preflight", "install", "apply", "postflight"} {
		if _, ok := names[want]; !ok {
			t.Fatalf("missing command %q", want)
		}
	}
}

func TestWriteCommandsRequireForceAndDiskID(t *testing.T) {
	for _, cmd := range installpolicy.Commands() {
		if !cmd.Mutates {
			continue
		}
		if !cmd.RequiresForce || !cmd.RequiresTargetDiskID {
			t.Fatalf("mutating command %q missing force/disk requirements", cmd.Name)
		}
	}
}
