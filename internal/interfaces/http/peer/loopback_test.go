package peer_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/peer"
)

func TestIsLoopback(t *testing.T) {
	allowed := []string{
		"127.0.0.1:8080",
		"127.0.0.2:1234",
		"[::1]:8080",
	}
	for _, addr := range allowed {
		if !peer.IsLoopback(addr) {
			t.Fatalf("expected loopback for %q", addr)
		}
	}

	rejected := []string{
		"",
		"192.168.1.10:8080",
		"8.8.8.8:443",
		"[fd00::1]:8080",
		"[fe80::1]:8080",
		"127.0.0.1",
		"malformed",
	}
	for _, addr := range rejected {
		if peer.IsLoopback(addr) {
			t.Fatalf("expected non-loopback for %q", addr)
		}
	}
}
