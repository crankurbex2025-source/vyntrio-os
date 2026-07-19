package installtarget_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installtarget"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

func eligibleInput(kernelName string) installtarget.RawDeviceInput {
	return installtarget.RawDeviceInput{
		KernelName: kernelName,
		SizeBytes:  installpreflight.MinInstallSizeBytes,
		SizeKnown:  true,
	}
}

func diskID(t *testing.T, kernelName string) string {
	t.Helper()
	return storageinventory.Classify(storageinventory.RawDevice{
		KernelName: kernelName,
		SizeBytes:  installpreflight.MinInstallSizeBytes,
		SizeKnown:  true,
	}).ID
}

func testApplier(t *testing.T, devices []installtarget.RawDeviceInput, partitions map[string][]string, fsTypes map[string]string) installtarget.Applier {
	t.Helper()
	mounter := &installtarget.StubMountController{}
	return installtarget.Applier{
		Reader:     installtarget.StubReader{Devices: devices},
		Mounter:    mounter,
		Prober:     installtarget.StubFSProber{Types: fsTypes},
		Partitions: installtarget.StubPartitionLister{Partitions: partitions},
	}
}

func TestPrepareSuccessSinglePartition(t *testing.T) {
	id := diskID(t, "sdb")
	mountRoot := t.TempDir()
	applier := testApplier(t,
		[]installtarget.RawDeviceInput{eligibleInput("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{"/dev/sdb1": "ext4"},
	)

	session, outcome, err := applier.Prepare(installtarget.ApplyRequest{
		DiskID:    id,
		StateDir:  t.TempDir(),
		MountRoot: mountRoot,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Prepare() error: %v", err)
	}
	if !outcome.Mounted {
		t.Fatal("expected mounted outcome")
	}
	if err := session.Unmount(); err != nil {
		t.Fatalf("Unmount() error: %v", err)
	}
}

func TestPrepareRequiresForce(t *testing.T) {
	id := diskID(t, "sdb")
	applier := testApplier(t,
		[]installtarget.RawDeviceInput{eligibleInput("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{"/dev/sdb1": "ext4"},
	)
	_, _, err := applier.Prepare(installtarget.ApplyRequest{DiskID: id, Force: false})
	if !errors.Is(err, installtarget.ErrForceRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestPrepareRejectsMountedTarget(t *testing.T) {
	id := diskID(t, "sdb")
	raw := eligibleInput("sdb")
	raw.Mounted = true
	applier := testApplier(t, []installtarget.RawDeviceInput{raw}, nil, nil)
	_, _, err := applier.Prepare(installtarget.ApplyRequest{DiskID: id, Force: true})
	if !errors.Is(err, installtarget.ErrTargetMounted) {
		t.Fatalf("err = %v", err)
	}
}

func TestPrepareRejectsRootDisk(t *testing.T) {
	id := diskID(t, "sda")
	raw := eligibleInput("sda")
	raw.RootDisk = true
	applier := testApplier(t, []installtarget.RawDeviceInput{raw}, nil, nil)
	_, _, err := applier.Prepare(installtarget.ApplyRequest{DiskID: id, Force: true})
	if !errors.Is(err, installtarget.ErrTargetNotEligible) {
		t.Fatalf("err = %v", err)
	}
}

func TestPrepareRejectsAmbiguousLayout(t *testing.T) {
	id := diskID(t, "sdb")
	applier := testApplier(t,
		[]installtarget.RawDeviceInput{eligibleInput("sdb")},
		map[string][]string{"sdb": {"sdb1", "sdb2"}},
		map[string]string{"/dev/sdb1": "ext4", "/dev/sdb2": "xfs"},
	)
	_, _, err := applier.Prepare(installtarget.ApplyRequest{DiskID: id, Force: true})
	if !errors.Is(err, installtarget.ErrAmbiguousTargetLayout) {
		t.Fatalf("err = %v", err)
	}
}

func TestPrepareRejectsUnsupportedTargetState(t *testing.T) {
	id := diskID(t, "sdb")
	applier := testApplier(t,
		[]installtarget.RawDeviceInput{eligibleInput("sdb")},
		map[string][]string{"sdb": {"sdb1"}},
		map[string]string{},
	)
	_, _, err := applier.Prepare(installtarget.ApplyRequest{DiskID: id, Force: true})
	if !errors.Is(err, installtarget.ErrUnsupportedTargetState) {
		t.Fatalf("err = %v", err)
	}
}

func TestWriteMutationRecord(t *testing.T) {
	stateDir := t.TempDir()
	path, err := installtarget.WriteMutationRecord(stateDir, installtarget.ApplyOutcome{
		DiskID:     "disk-test",
		MountPoint: "/run/vyntrio-install/mnt/disk-test",
		FSType:     "ext4",
		Candidate:  installtarget.MountCandidate{DevicePath: "/dev/sdb1", FSType: "ext4"},
	}, installtarget.StatusApplied, 6, []string{"usr/bin/vyntrio-api"})
	if err != nil {
		t.Fatalf("WriteMutationRecord() error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("record missing: %v", err)
	}
	if filepath.Base(path) != installtarget.MutationRecordName {
		t.Fatalf("record name = %s", filepath.Base(path))
	}
}
