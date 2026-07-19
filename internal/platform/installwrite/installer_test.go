package installwrite_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installwrite"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

type stubReader struct {
	devices []storageinventory.RawDevice
}

func (s stubReader) ListBlockDevices(string) ([]storageinventory.RawDevice, error) {
	return s.devices, nil
}

func eligibleDevice(kernelName string) storageinventory.RawDevice {
	size := installpreflight.MinInstallSizeBytes
	return storageinventory.RawDevice{
		KernelName: kernelName,
		SizeBytes:  size,
		SizeKnown:  true,
	}
}

func deviceID(t *testing.T, kernelName string) string {
	t.Helper()
	return storageinventory.Classify(eligibleDevice(kernelName)).ID
}

func installerWithDevice(t *testing.T, sandbox string, devices []storageinventory.RawDevice) installwrite.Installer {
	t.Helper()
	return installwrite.Installer{
		Checker: installpreflight.Checker{
			Collector: storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
				Reader: stubReader{devices: devices},
			}),
		},
	}
}

func writeEnvelope(t *testing.T) string {
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
		if err := os.WriteFile(path, []byte("payload-"+rel), 0o644); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}
	}
	return root
}

func TestInstallSuccessWithEnvelope(t *testing.T) {
	sandbox := t.TempDir()
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	installer := installerWithDevice(t, sandbox, []storageinventory.RawDevice{eligibleDevice("sdb")})

	outcome, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: id,
		SandboxRoot:  sandbox,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if outcome.PayloadsCopied != 6 {
		t.Fatalf("payloads = %d", outcome.PayloadsCopied)
	}
	if _, err := os.Stat(filepath.Join(outcome.TargetRoot, "usr/bin/vyntrio-api")); err != nil {
		t.Fatalf("missing copied payload: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outcome.TargetRoot, installwrite.InstallRecordName)); err != nil {
		t.Fatalf("missing install record: %v", err)
	}
}

func TestInstallRequiresForce(t *testing.T) {
	sandbox := t.TempDir()
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	installer := installerWithDevice(t, sandbox, []storageinventory.RawDevice{eligibleDevice("sdb")})

	_, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: id,
		SandboxRoot:  sandbox,
		EnvelopeRoot: envelope,
		Force:        false,
	})
	if err != installwrite.ErrForceRequired {
		t.Fatalf("err = %v", err)
	}
}

func TestInstallRejectsRootDiskTarget(t *testing.T) {
	sandbox := t.TempDir()
	envelope := writeEnvelope(t)
	raw := eligibleDevice("sda")
	raw.RootDisk = true
	installer := installerWithDevice(t, sandbox, []storageinventory.RawDevice{raw})
	id := storageinventory.Classify(raw).ID

	_, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: id,
		SandboxRoot:  sandbox,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstallRejectsUnsafeTargetRoot(t *testing.T) {
	sandbox := t.TempDir()
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	installer := installerWithDevice(t, sandbox, []storageinventory.RawDevice{eligibleDevice("sdb")})

	_, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: id,
		SandboxRoot:  sandbox,
		TargetRoot:   t.TempDir(),
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if !errors.Is(err, installwrite.ErrUnsafeTargetRoot) {
		t.Fatalf("err = %v", err)
	}
}

func TestInstallRequiresArtifactSource(t *testing.T) {
	sandbox := t.TempDir()
	id := deviceID(t, "sdb")
	installer := installerWithDevice(t, sandbox, []storageinventory.RawDevice{eligibleDevice("sdb")})

	_, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: id,
		SandboxRoot:  sandbox,
		Force:        true,
	})
	if err != installwrite.ErrArtifactSourceRequired {
		t.Fatalf("err = %v", err)
	}
}

func TestInstallRejectsAmbiguousTarget(t *testing.T) {
	sandbox := t.TempDir()
	envelope := writeEnvelope(t)
	raw := eligibleDevice("sde")
	raw.SizeKnown = false
	installer := installerWithDevice(t, sandbox, []storageinventory.RawDevice{raw})
	id := storageinventory.Classify(raw).ID

	_, err := installer.Install(context.Background(), installwrite.Request{
		TargetDiskID: id,
		SandboxRoot:  sandbox,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
