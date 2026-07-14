package netpresence_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/netpresence"
)

type stubInterfacesReader struct {
	flags []netpresence.InterfaceFlags
	err   error
}

func (s stubInterfacesReader) Interfaces() ([]netpresence.InterfaceFlags, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.flags, nil
}

func TestCollectorReaderErrorReturnsUnavailable(t *testing.T) {
	collector := netpresence.NewCollector(netpresence.CollectorDeps{
		Reader: stubInterfacesReader{err: errors.New("enumeration failed")},
	})
	got := collector.Collect(context.Background())
	if got.Status != netpresence.StatusUnavailable {
		t.Fatalf("status = %q, want unavailable", got.Status)
	}
}

func TestCollectorEligibleInterfaceReturnsAvailable(t *testing.T) {
	collector := netpresence.NewCollector(netpresence.CollectorDeps{
		Reader: stubInterfacesReader{flags: []netpresence.InterfaceFlags{{
			Up:              true,
			HasHardwareAddr: true,
		}}},
	})
	got := collector.Collect(context.Background())
	if got.Status != netpresence.StatusAvailable {
		t.Fatalf("status = %q, want available", got.Status)
	}
}

func TestCollectorNoEligibleInterfacesReturnsUnknown(t *testing.T) {
	collector := netpresence.NewCollector(netpresence.CollectorDeps{
		Reader: stubInterfacesReader{flags: []netpresence.InterfaceFlags{{
			Loopback:        true,
			Up:              true,
			HasHardwareAddr: true,
		}}},
	})
	got := collector.Collect(context.Background())
	if got.Status != netpresence.StatusUnknown {
		t.Fatalf("status = %q, want unknown", got.Status)
	}
}

func TestNetworkJSONContainsOnlyStatus(t *testing.T) {
	data, err := json.Marshal(netpresence.Network{Status: netpresence.StatusAvailable})
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	body := strings.ToLower(string(data))
	for _, forbidden := range []string{
		`"interface"`, `"name"`, `"mac"`, `"address"`, `"ip"`, `"route"`,
		`"gateway"`, `"dns"`, `"mtu"`, `"index"`, `"count"`, `"error"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("json contained forbidden field %q: %s", forbidden, data)
		}
	}
	if body != `{"status":"available"}` {
		t.Fatalf("json = %s, want status-only object", data)
	}
}
