package netpresence_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/netpresence"
)

func TestIsEligibleRejectsLoopback(t *testing.T) {
	if netpresence.IsEligible(netpresence.InterfaceFlags{
		Loopback:        true,
		Up:              true,
		HasHardwareAddr: true,
	}) {
		t.Fatal("expected loopback interface to be ineligible")
	}
}

func TestIsEligibleRejectsDownInterface(t *testing.T) {
	if netpresence.IsEligible(netpresence.InterfaceFlags{Up: false, HasHardwareAddr: true}) {
		t.Fatal("expected down interface to be ineligible")
	}
}

func TestIsEligibleRejectsEmptyHardwareAddress(t *testing.T) {
	if netpresence.IsEligible(netpresence.InterfaceFlags{Up: true, HasHardwareAddr: false}) {
		t.Fatal("expected empty hardware address to be ineligible")
	}
}

func TestClassifyAvailableWithEligibleInterface(t *testing.T) {
	got := netpresence.Classify([]netpresence.InterfaceFlags{{
		Up:              true,
		HasHardwareAddr: true,
	}})
	if got.Status != netpresence.StatusAvailable {
		t.Fatalf("status = %q, want available", got.Status)
	}
}

func TestClassifyUnknownWithNoEligibleInterfaces(t *testing.T) {
	got := netpresence.Classify([]netpresence.InterfaceFlags{
		{Loopback: true, Up: true, HasHardwareAddr: true},
		{Up: false, HasHardwareAddr: true},
	})
	if got.Status != netpresence.StatusUnknown {
		t.Fatalf("status = %q, want unknown", got.Status)
	}
}
