package storageinventory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

type stubBlockDeviceReader struct {
	devices []storageinventory.RawDevice
	err     error
}

func (s stubBlockDeviceReader) ListBlockDevices(string) ([]storageinventory.RawDevice, error) {
	return s.devices, s.err
}

func TestCollectorReaderErrorReturnsUnavailable(t *testing.T) {
	collector := storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
		Reader: stubBlockDeviceReader{err: errors.New("discovery failed")},
	})
	got := collector.Collect(context.Background())
	if got.Status != storageinventory.StatusUnavailable {
		t.Fatalf("status = %q, want unavailable", got.Status)
	}
	if got.Devices != nil {
		t.Fatalf("devices = %v, want nil", got.Devices)
	}
}

func TestCollectorClassifiesDiscoveredDevices(t *testing.T) {
	size := uint64(2_000_000_000_000)
	rot := false
	collector := storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
		Reader: stubBlockDeviceReader{
			devices: []storageinventory.RawDevice{
				{
					KernelName: "sda",
					SizeBytes:  size,
					SizeKnown:  true,
					RootDisk:   true,
				},
				{
					KernelName: "sdb",
					SizeBytes:  size,
					SizeKnown:  true,
					Rotational: &rot,
				},
			},
		},
	})
	got := collector.Collect(context.Background())
	if got.Status != storageinventory.StatusOK {
		t.Fatalf("status = %q, want ok", got.Status)
	}
	if len(got.Devices) != 2 {
		t.Fatalf("len(devices) = %d, want 2", len(got.Devices))
	}
	if got.Devices[0].Eligibility != storageinventory.EligibilityExcluded {
		t.Fatalf("root eligibility = %q", got.Devices[0].Eligibility)
	}
	if got.Devices[1].Eligibility != storageinventory.EligibilityEligible {
		t.Fatalf("candidate eligibility = %q", got.Devices[1].Eligibility)
	}
}

func TestCollectorAmbiguousDeviceMarkedUnknown(t *testing.T) {
	collector := storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
		Reader: stubBlockDeviceReader{
			devices: []storageinventory.RawDevice{
				{KernelName: "", IdentityAmbiguous: true},
			},
		},
	})
	got := collector.Collect(context.Background())
	if len(got.Devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(got.Devices))
	}
	if got.Devices[0].Eligibility != storageinventory.EligibilityUnknown {
		t.Fatalf("eligibility = %q, want unknown", got.Devices[0].Eligibility)
	}
}
