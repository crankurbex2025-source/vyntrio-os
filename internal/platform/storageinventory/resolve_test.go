package storageinventory_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

func TestLookupRawDeviceSuccess(t *testing.T) {
	raw := storageinventory.RawDevice{KernelName: "sdb", SizeBytes: 1 << 30, SizeKnown: true}
	id := storageinventory.Classify(raw).ID
	found, count, err := storageinventory.LookupRawDevice([]storageinventory.RawDevice{raw}, id)
	if err != nil {
		t.Fatalf("LookupRawDevice() error: %v", err)
	}
	if count != 1 || found.KernelName != "sdb" {
		t.Fatalf("found = %+v count = %d", found, count)
	}
}

func TestLookupRawDeviceAmbiguous(t *testing.T) {
	raw := storageinventory.RawDevice{KernelName: "sdb", SizeBytes: 1 << 30, SizeKnown: true}
	id := storageinventory.Classify(raw).ID
	_, count, err := storageinventory.LookupRawDevice([]storageinventory.RawDevice{raw, raw}, id)
	if err == nil || count != 2 {
		t.Fatalf("expected ambiguous lookup, count=%d err=%v", count, err)
	}
}

func TestLookupRawDeviceNotFound(t *testing.T) {
	_, _, err := storageinventory.LookupRawDevice(nil, "disk-missing")
	if err == nil {
		t.Fatal("expected error")
	}
}
