package storageinventory_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

func TestClassifyEligibleDisk(t *testing.T) {
	size := uint64(1_000_000_000_000)
	rot := false
	raw := storageinventory.RawDevice{
		KernelName: "sdb",
		SizeBytes:  size,
		SizeKnown:  true,
		Rotational: &rot,
	}
	got := storageinventory.Classify(raw)
	if got.Eligibility != storageinventory.EligibilityEligible {
		t.Fatalf("eligibility = %q, want eligible", got.Eligibility)
	}
	if len(got.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", got.Reasons)
	}
	if got.ID == "" || got.ID == "disk-unknown" {
		t.Fatalf("id = %q, want stable opaque id", got.ID)
	}
	if got.SizeBytes == nil || *got.SizeBytes != size {
		t.Fatalf("size_bytes = %v", got.SizeBytes)
	}
}

func TestClassifyExclusionReasons(t *testing.T) {
	size := uint64(500_000_000_000)
	cases := []struct {
		name        string
		raw         storageinventory.RawDevice
		wantReasons []string
	}{
		{
			name: "root_disk",
			raw: storageinventory.RawDevice{
				KernelName: "sda",
				SizeBytes:  size,
				SizeKnown:  true,
				RootDisk:   true,
			},
			wantReasons: []string{storageinventory.ReasonRootDisk},
		},
		{
			name: "state_filesystem",
			raw: storageinventory.RawDevice{
				KernelName: "nvme0n1",
				SizeBytes:  size,
				SizeKnown:  true,
				StateDisk:  true,
			},
			wantReasons: []string{storageinventory.ReasonStateFilesystem},
		},
		{
			name: "removable",
			raw: storageinventory.RawDevice{
				KernelName: "sdc",
				SizeBytes:  size,
				SizeKnown:  true,
				Removable:  true,
			},
			wantReasons: []string{storageinventory.ReasonRemovable},
		},
		{
			name: "read_only",
			raw: storageinventory.RawDevice{
				KernelName: "sdd",
				SizeBytes:  size,
				SizeKnown:  true,
				ReadOnly:   true,
			},
			wantReasons: []string{storageinventory.ReasonReadOnly},
		},
		{
			name: "mounted_in_use",
			raw: storageinventory.RawDevice{
				KernelName: "sde",
				SizeBytes:  size,
				SizeKnown:  true,
				Mounted:    true,
			},
			wantReasons: []string{storageinventory.ReasonMountedInUse},
		},
		{
			name: "unsupported_filesystem",
			raw: storageinventory.RawDevice{
				KernelName: "sdf",
				SizeBytes:  size,
				SizeKnown:  true,
				FSTypes:    []string{"ntfs"},
			},
			wantReasons: []string{storageinventory.ReasonUnsupportedFilesystem},
		},
		{
			name: "install_media",
			raw: storageinventory.RawDevice{
				KernelName: "sr0",
				SizeBytes:  size,
				SizeKnown:  true,
				Optical:    true,
			},
			wantReasons: []string{storageinventory.ReasonInstallMedia},
		},
		{
			name: "virtual_device",
			raw: storageinventory.RawDevice{
				KernelName: "sdb",
				SizeBytes:  size,
				SizeKnown:  true,
				Virtual:    true,
			},
			wantReasons: []string{storageinventory.ReasonVirtualDevice},
		},
		{
			name: "root_disk_skips_mounted_in_use",
			raw: storageinventory.RawDevice{
				KernelName: "sda",
				SizeBytes:  size,
				SizeKnown:  true,
				RootDisk:   true,
				Mounted:    true,
			},
			wantReasons: []string{storageinventory.ReasonRootDisk},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := storageinventory.Classify(tc.raw)
			if got.Eligibility != storageinventory.EligibilityExcluded {
				t.Fatalf("eligibility = %q, want excluded", got.Eligibility)
			}
			if len(got.Reasons) != len(tc.wantReasons) {
				t.Fatalf("reasons = %v, want %v", got.Reasons, tc.wantReasons)
			}
			for i, reason := range tc.wantReasons {
				if got.Reasons[i] != reason {
					t.Fatalf("reasons[%d] = %q, want %q", i, got.Reasons[i], reason)
				}
			}
		})
	}
}

func TestClassifyAmbiguousIdentityFailClosed(t *testing.T) {
	cases := []struct {
		name string
		raw  storageinventory.RawDevice
	}{
		{
			name: "empty_kernel_name",
			raw:  storageinventory.RawDevice{KernelName: ""},
		},
		{
			name: "identity_flag",
			raw: storageinventory.RawDevice{
				KernelName:        "sdb",
				SizeBytes:         1_000,
				SizeKnown:         true,
				IdentityAmbiguous: true,
			},
		},
		{
			name: "unknown_size",
			raw: storageinventory.RawDevice{
				KernelName: "sdb",
				SizeKnown:  false,
			},
		},
		{
			name: "zero_size",
			raw: storageinventory.RawDevice{
				KernelName: "sdb",
				SizeBytes:  0,
				SizeKnown:  true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := storageinventory.Classify(tc.raw)
			if got.Eligibility != storageinventory.EligibilityUnknown {
				t.Fatalf("eligibility = %q, want unknown", got.Eligibility)
			}
			if !containsReason(got.Reasons, storageinventory.ReasonAmbiguousIdentity) {
				t.Fatalf("reasons = %v, want ambiguous_identity", got.Reasons)
			}
			if tc.name == "empty_kernel_name" || tc.name == "identity_flag" {
				if got.ID != "disk-unknown" {
					t.Fatalf("id = %q, want disk-unknown", got.ID)
				}
			} else if got.ID == "" || got.ID == "disk-unknown" {
				t.Fatalf("id = %q, want stable opaque id", got.ID)
			}
		})
	}
}

func containsReason(reasons []string, want string) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}

func TestClassifySupportedFilesystemDoesNotExclude(t *testing.T) {
	size := uint64(1_000_000_000)
	raw := storageinventory.RawDevice{
		KernelName: "sdb",
		SizeBytes:  size,
		SizeKnown:  true,
		FSTypes:    []string{"ext4", "xfs"},
	}
	got := storageinventory.Classify(raw)
	if got.Eligibility != storageinventory.EligibilityEligible {
		t.Fatalf("eligibility = %q, want eligible", got.Eligibility)
	}
}

func TestClassifyStableDeviceID(t *testing.T) {
	size := uint64(1_000_000_000)
	raw := storageinventory.RawDevice{
		KernelName: "sdb",
		SizeBytes:  size,
		SizeKnown:  true,
	}
	first := storageinventory.Classify(raw)
	second := storageinventory.Classify(raw)
	if first.ID != second.ID {
		t.Fatalf("ids differ: %q vs %q", first.ID, second.ID)
	}
}
