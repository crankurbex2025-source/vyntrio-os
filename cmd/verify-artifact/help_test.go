package main

import (
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpolicy"
)

func TestVerifyArtifactUsageContract(t *testing.T) {
	text := installpolicy.VerifyArtifactUsageText()
	for _, marker := range []string{
		"USB-first",
		"NOT the primary install journey",
		"preflight",
		"install",
		"apply",
		"Does NOT create USB/ISO",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("usage missing %q", marker)
		}
	}
}
