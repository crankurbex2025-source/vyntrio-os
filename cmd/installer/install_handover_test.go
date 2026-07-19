package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpostflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installwrite"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

func TestRunInstallWritesHandoverRecord(t *testing.T) {
	sandbox := t.TempDir()
	envelope := writeTestEnvelope(t)
	size := installpreflight.MinInstallSizeBytes
	device := storageinventory.Classify(storageinventory.RawDevice{
		KernelName: "sdb",
		SizeBytes:  size,
		SizeKnown:  true,
	})
	installer := installwrite.Installer{
		Checker: installpreflight.Checker{
			Collector: storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
				Reader: stubReader{devices: []storageinventory.RawDevice{{
					KernelName: "sdb",
					SizeBytes:  size,
					SizeKnown:  true,
				}}},
			}),
		},
	}
	outcome, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: device.ID,
		SandboxRoot:  sandbox,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	handover := installpostflight.Build(installpostflight.InputFromInstall("", envelope, outcome, installpostflight.WriteSucceeded))
	if _, err := installpostflight.WriteRecord(outcome.TargetRoot, handover); err != nil {
		t.Fatalf("WriteRecord() error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outcome.TargetRoot, installpostflight.HandoverRecordName)); err != nil {
		t.Fatalf("missing handover record: %v", err)
	}
}

type stubReader struct {
	devices []storageinventory.RawDevice
}

func (s stubReader) ListBlockDevices(string) ([]storageinventory.RawDevice, error) {
	return s.devices, nil
}

func writeTestEnvelope(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ENVELOPE.txt"), []byte("media_role: install\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	for _, rel := range installpreflight.RequiredPayloadRelativePaths() {
		path := filepath.Join(root, "payload", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error: %v", err)
		}
		if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}
	}
	return root
}
