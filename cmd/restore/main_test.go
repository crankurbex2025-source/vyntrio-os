package main

import (
	"os"
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

func TestRunValidateRequiresRoot(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root environment skips non-root guard test")
	}
	if run([]string{"validate", "vyntrio-backup-v1_20260715T100000.000000000Z.tar"}) != 1 {
		t.Fatal("expected failure for non-root validate")
	}
}

func TestRunRestoreRequiresForceFlag(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root environment covered by package runner tests")
	}
	if run([]string{"vyntrio-backup-v1_20260715T100000.000000000Z.tar"}) != 1 {
		t.Fatal("expected failure without --force")
	}
}
