package main

import (
	"testing"
)

func TestRunUsage(t *testing.T) {
	if run(nil) != 2 {
		t.Fatal("expected usage exit code 2")
	}
	if run([]string{"help"}) != 0 {
		t.Fatal("expected help exit code 0")
	}
}

func TestRunPreflightRequiresTargetDiskID(t *testing.T) {
	if run([]string{"preflight"}) != 1 {
		t.Fatal("expected failure without target disk id")
	}
}

func TestRunPreflightRejectsInvalidMinSize(t *testing.T) {
	if run([]string{"preflight", "--target-disk-id", "disk-abc", "--min-size-bytes", "not-a-number"}) != 1 {
		t.Fatal("expected failure for invalid min size")
	}
}

func TestRunInstallRequiresForce(t *testing.T) {
	if run([]string{"install", "--target-disk-id", "disk-abc", "--envelope-root", "/tmp/env"}) != 1 {
		t.Fatal("expected failure without --force")
	}
}

func TestRunInstallRequiresArtifactSource(t *testing.T) {
	if run([]string{"install", "--target-disk-id", "disk-abc", "--force"}) != 1 {
		t.Fatal("expected failure without artifact source")
	}
}

func TestRunApplyRequiresForce(t *testing.T) {
	if run([]string{"apply", "--target-disk-id", "disk-abc", "--envelope-root", "/tmp/env"}) != 1 {
		t.Fatal("expected failure without --force")
	}
}

func TestRunApplyRequiresArtifactSource(t *testing.T) {
	if run([]string{"apply", "--target-disk-id", "disk-abc", "--force"}) != 1 {
		t.Fatal("expected failure without artifact source")
	}
}

func TestRunPostflightRequiresTargetRoot(t *testing.T) {
	if run([]string{"postflight"}) != 1 {
		t.Fatal("expected failure without target root")
	}
}
