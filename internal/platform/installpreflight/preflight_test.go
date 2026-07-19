package installpreflight_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

type stubReader struct {
	devices []storageinventory.RawDevice
	err     error
}

func (s stubReader) ListBlockDevices(string) ([]storageinventory.RawDevice, error) {
	return s.devices, s.err
}

func checkerWithDevices(t *testing.T, devices []storageinventory.RawDevice) installpreflight.Checker {
	t.Helper()
	return installpreflight.Checker{
		Collector: storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
			Reader: stubReader{devices: devices},
		}),
	}
}

func deviceID(t *testing.T, kernelName string) string {
	t.Helper()
	device := storageinventory.Classify(storageinventory.RawDevice{
		KernelName: kernelName,
		SizeBytes:  installpreflight.MinInstallSizeBytes,
		SizeKnown:  true,
	})
	return device.ID
}

func eligibleDevice(kernelName string) storageinventory.RawDevice {
	size := installpreflight.MinInstallSizeBytes
	rot := false
	return storageinventory.RawDevice{
		KernelName: kernelName,
		SizeBytes:  size,
		SizeKnown:  true,
		Rotational: &rot,
	}
}

func TestTargetEligiblePasses(t *testing.T) {
	checker := checkerWithDevices(t, []storageinventory.RawDevice{
		eligibleDevice("sdb"),
	})
	id := deviceID(t, "sdb")

	result, err := checker.Run(context.Background(), installpreflight.TargetRequest{
		DiskID: id,
	}, installpreflight.MediaRequest{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.Target.Status != installpreflight.TargetEligible {
		t.Fatalf("status = %q", result.Target.Status)
	}
}

func TestTargetMissingSelectionFails(t *testing.T) {
	checker := installpreflight.NewChecker("/var/lib/vyntrio")
	_, err := checker.Run(context.Background(), installpreflight.TargetRequest{}, installpreflight.MediaRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTargetNotFoundFails(t *testing.T) {
	checker := checkerWithDevices(t, []storageinventory.RawDevice{
		eligibleDevice("sdb"),
	})
	_, err := checker.Run(context.Background(), installpreflight.TargetRequest{
		DiskID: "disk-deadbeefdead",
	}, installpreflight.MediaRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTargetRootDiskFails(t *testing.T) {
	raw := eligibleDevice("sda")
	raw.RootDisk = true
	checker := checkerWithDevices(t, []storageinventory.RawDevice{raw})
	id := deviceID(t, "sda")

	result, err := checker.Run(context.Background(), installpreflight.TargetRequest{DiskID: id}, installpreflight.MediaRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Target.Status != installpreflight.TargetExcluded {
		t.Fatalf("status = %q", result.Target.Status)
	}
}

func TestTargetMountedInUseFails(t *testing.T) {
	raw := eligibleDevice("sdc")
	raw.Mounted = true
	checker := checkerWithDevices(t, []storageinventory.RawDevice{raw})
	id := deviceID(t, "sdc")

	result, err := checker.Run(context.Background(), installpreflight.TargetRequest{DiskID: id}, installpreflight.MediaRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Target.Status != installpreflight.TargetExcluded {
		t.Fatalf("status = %q", result.Target.Status)
	}
}

func TestTargetInsufficientSizeFails(t *testing.T) {
	size := uint64(1024 * 1024)
	raw := storageinventory.RawDevice{
		KernelName: "sdd",
		SizeBytes:  size,
		SizeKnown:  true,
	}
	checker := checkerWithDevices(t, []storageinventory.RawDevice{raw})
	id := deviceID(t, "sdd")

	result, err := checker.Run(context.Background(), installpreflight.TargetRequest{DiskID: id}, installpreflight.MediaRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Target.Status != installpreflight.TargetExcluded {
		t.Fatalf("status = %q", result.Target.Status)
	}
}

func TestTargetAmbiguousIdentityFails(t *testing.T) {
	raw := eligibleDevice("sde")
	raw.SizeKnown = false
	checker := checkerWithDevices(t, []storageinventory.RawDevice{raw})
	id := deviceID(t, "sde")

	result, err := checker.Run(context.Background(), installpreflight.TargetRequest{DiskID: id}, installpreflight.MediaRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Target.Status != installpreflight.TargetUnknown {
		t.Fatalf("status = %q", result.Target.Status)
	}
}

func TestMediaEnvelopeValidPasses(t *testing.T) {
	root := writeEnvelope(t)
	checker := checkerWithDevices(t, []storageinventory.RawDevice{eligibleDevice("sdb")})
	id := deviceID(t, "sdb")

	result, err := checker.Run(context.Background(),
		installpreflight.TargetRequest{DiskID: id},
		installpreflight.MediaRequest{EnvelopeRoot: root},
	)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.Media.EnvelopeStatus != installpreflight.MediaOK {
		t.Fatalf("envelope = %q", result.Media.EnvelopeStatus)
	}
}

func TestMediaEnvelopeMissingFails(t *testing.T) {
	checker := checkerWithDevices(t, []storageinventory.RawDevice{eligibleDevice("sdb")})
	id := deviceID(t, "sdb")

	_, err := checker.Run(context.Background(),
		installpreflight.TargetRequest{DiskID: id},
		installpreflight.MediaRequest{EnvelopeRoot: t.TempDir() + "/missing"},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMediaExcludedPayloadFails(t *testing.T) {
	root := writeEnvelope(t)
	payloadRoot := filepath.Join(root, "payload")
	if err := os.WriteFile(filepath.Join(payloadRoot, "state.db"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	checker := checkerWithDevices(t, []storageinventory.RawDevice{eligibleDevice("sdb")})
	id := deviceID(t, "sdb")

	_, err := checker.Run(context.Background(),
		installpreflight.TargetRequest{DiskID: id},
		installpreflight.MediaRequest{EnvelopeRoot: root},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeEnvelope(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ENVELOPE.txt"), []byte("media_role: install\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	payloadRoot := filepath.Join(root, "payload")
	for _, rel := range []string{
		"usr/bin/vyntrio-api",
		"usr/bin/vyntrio-backup",
		"etc/systemd/system/vyntrio-api.service",
		"usr/lib/sysusers.d/vyntrio.conf",
		"etc/tmpfiles.d/vyntrio.conf",
		"etc/vyntrio/config.toml",
	} {
		path := filepath.Join(payloadRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error: %v", err)
		}
		if err := os.WriteFile(path, []byte("fixture"), 0o644); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}
	}
	return root
}
