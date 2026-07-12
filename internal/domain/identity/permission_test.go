package identity_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

func TestParsePermissionValid(t *testing.T) {
	for _, perm := range identity.AllPermissions {
		got, err := identity.ParsePermission(perm.String())
		if err != nil {
			t.Fatalf("ParsePermission(%q) error: %v", perm, err)
		}
		if got != perm {
			t.Fatalf("ParsePermission(%q) = %q", perm, got)
		}
	}
}

func TestParsePermissionInvalid(t *testing.T) {
	_, err := identity.ParsePermission("secrets:dump")
	if err == nil {
		t.Fatal("expected error for unknown permission")
	}
}
