package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpolicy"
)

func TestRunWorkflowPrintsPolicy(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error: %v", err)
	}
	os.Stdout = w
	code := run([]string{"workflow"})
	_ = w.Close()
	os.Stdout = old
	if code != 0 {
		t.Fatalf("workflow exit = %d", code)
	}
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	text := buf.String()
	for _, marker := range installpolicy.RequiredWorkflowMarkers() {
		if !strings.Contains(text, marker) {
			t.Fatalf("workflow output missing %q", marker)
		}
	}
}

func TestUsageHelpContract(t *testing.T) {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error: %v", err)
	}
	os.Stderr = w
	_ = run([]string{"help"})
	_ = w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	text := buf.String()
	for _, marker := range installpolicy.RequiredUsageMarkers() {
		if !strings.Contains(text, marker) {
			t.Fatalf("help missing %q", marker)
		}
	}
}
