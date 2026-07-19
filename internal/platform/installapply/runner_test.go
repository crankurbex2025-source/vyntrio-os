package installapply_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installapply"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installtarget"
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
	return storageinventory.RawDevice{
		KernelName: kernelName,
		SizeBytes:  installpreflight.MinInstallSizeBytes,
		SizeKnown:  true,
	}
}

func deviceID(t *testing.T, kernelName string) string {
	t.Helper()
	return storageinventory.Classify(eligibleDevice(kernelName)).ID
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

func runnerWithDevice(
	t *testing.T,
	devices []storageinventory.RawDevice,
	partitions map[string][]string,
	fsTypes map[string]string,
	mounter installtarget.MountController,
) installapply.Runner {
	t.Helper()
	readerDevices := make([]installtarget.RawDeviceInput, 0, len(devices))
	for _, device := range devices {
		readerDevices = append(readerDevices, installtarget.RawDeviceInput{
			KernelName:        device.KernelName,
			SizeBytes:         device.SizeBytes,
			SizeKnown:         device.SizeKnown,
			Removable:         device.Removable,
			ReadOnly:          device.ReadOnly,
			Virtual:           device.Virtual,
			Optical:           device.Optical,
			Mounted:           device.Mounted,
			RootDisk:          device.RootDisk,
			StateDisk:         device.StateDisk,
			IdentityAmbiguous: device.IdentityAmbiguous,
		})
	}
	if mounter == nil {
		mounter = &installtarget.StubMountController{}
	}
	return installapply.Runner{
		Checker: installpreflight.Checker{
			Collector: storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
				Reader: stubReader{devices: devices},
			}),
		},
		Applier: installtarget.Applier{
			Reader:     installtarget.StubReader{Devices: readerDevices},
			Mounter:    mounter,
			Prober:     installtarget.StubFSProber{Types: fsTypes},
			Partitions: installtarget.StubPartitionLister{Partitions: partitions},
		},
	}
}

func TestApplySuccessExistingPartition(t *testing.T) {
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	mountRoot := t.TempDir()
	stateDir := t.TempDir()
	runner := runnerWithDevice(t,
		[]storageinventory.RawDevice{eligibleDevice("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{"/dev/sdb1": "ext4"},
		nil,
	)

	outcome, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
		MountRoot:    mountRoot,
		StateDir:     stateDir,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if outcome.PayloadsCopied != 6 {
		t.Fatalf("payloads = %d", outcome.PayloadsCopied)
	}
	if !outcome.HostBlockDeviceMutated {
		t.Fatal("expected host mutation")
	}
	if _, err := os.Stat(outcome.ApplyRecordPath); err != nil {
		t.Fatalf("apply record missing: %v", err)
	}
	if _, err := os.Stat(outcome.MutationRecordPath); err != nil {
		t.Fatalf("mutation record missing: %v", err)
	}
}

func TestApplyRequiresForce(t *testing.T) {
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	runner := runnerWithDevice(t,
		[]storageinventory.RawDevice{eligibleDevice("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{"/dev/sdb1": "ext4"},
		nil,
	)
	_, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        false,
	})
	if !errors.Is(err, installapply.ErrForceRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyRejectsRootDisk(t *testing.T) {
	envelope := writeEnvelope(t)
	raw := eligibleDevice("sda")
	raw.RootDisk = true
	id := storageinventory.Classify(raw).ID
	runner := runnerWithDevice(t, []storageinventory.RawDevice{raw}, nil, nil, nil)
	_, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if !errors.Is(err, installtarget.ErrTargetNotEligible) && !errors.Is(err, installapply.ErrPreflightFailed) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyRejectsMountedDisk(t *testing.T) {
	envelope := writeEnvelope(t)
	raw := eligibleDevice("sdb")
	raw.Mounted = true
	id := storageinventory.Classify(raw).ID
	runner := runnerWithDevice(t, []storageinventory.RawDevice{raw}, nil, nil, nil)
	_, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, installapply.ErrPreflightFailed) && !errors.Is(err, installtarget.ErrTargetMounted) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyRejectsBlankDisk(t *testing.T) {
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	runner := runnerWithDevice(t,
		[]storageinventory.RawDevice{eligibleDevice("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{},
		nil,
	)
	_, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if !errors.Is(err, installtarget.ErrUnsupportedTargetState) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyRejectsAmbiguousLayout(t *testing.T) {
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	runner := runnerWithDevice(t,
		[]storageinventory.RawDevice{eligibleDevice("sdb")},
		map[string][]string{"sdb": {"sdb1", "sdb2"}},
		map[string]string{"/dev/sdb1": "ext4", "/dev/sdb2": "xfs"},
		nil,
	)
	_, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
	})
	if !errors.Is(err, installtarget.ErrAmbiguousTargetLayout) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyRollbackOnMountFailure(t *testing.T) {
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	stateDir := t.TempDir()
	mounter := &installtarget.StubMountController{
		MountFn: func(string, string, string) error {
			return errors.New("mount refused")
		},
	}
	runner := runnerWithDevice(t,
		[]storageinventory.RawDevice{eligibleDevice("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{"/dev/sdb1": "ext4"},
		mounter,
	)
	outcome, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
		StateDir:     stateDir,
	})
	if !errors.Is(err, installtarget.ErrMountFailed) {
		t.Fatalf("err = %v", err)
	}
	if outcome.ApplyRecordPath == "" {
		t.Fatal("expected apply record on mount failure")
	}
}

func TestApplyRollbackOnCopyFailure(t *testing.T) {
	envelope := writeEnvelope(t)
	id := deviceID(t, "sdb")
	mountRoot := t.TempDir()
	stateDir := t.TempDir()
	runner := runnerWithDevice(t,
		[]storageinventory.RawDevice{eligibleDevice("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{"/dev/sdb1": "ext4"},
		nil,
	)
	calls := 0
	runner.CopyPayload = func(entry installwrite.CopyEntry, targetRoot string) error {
		calls++
		if calls == 2 {
			return errors.New("injected copy failure")
		}
		return installwrite.CopyPayloadFile(entry, targetRoot)
	}
	outcome, err := runner.Run(context.Background(), installapply.Request{
		TargetDiskID: id,
		EnvelopeRoot: envelope,
		Force:        true,
		MountRoot:    mountRoot,
		StateDir:     stateDir,
	})
	if err == nil {
		t.Fatal("expected copy failure")
	}
	if outcome.PayloadsCopied != 1 {
		t.Fatalf("payloads copied = %d", outcome.PayloadsCopied)
	}
	if outcome.FailureStage != installapply.StagePayloadCopy {
		t.Fatalf("stage = %q", outcome.FailureStage)
	}
	if outcome.ApplyRecordPath == "" || outcome.MutationRecordPath == "" {
		t.Fatal("expected rollback records")
	}
	data, readErr := os.ReadFile(outcome.ApplyRecordPath)
	if readErr != nil {
		t.Fatalf("read apply record: %v", readErr)
	}
	if !contains(string(data), installapply.StatusRolledBack) {
		t.Fatalf("expected rolled_back status in %q", string(data))
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(needle) == 0 ||
		(len(haystack) > 0 && stringIndex(haystack, needle) >= 0))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
