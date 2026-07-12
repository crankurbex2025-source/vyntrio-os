package identity_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

func TestParseRoleValid(t *testing.T) {
	for _, role := range identity.AllRoles {
		got, err := identity.ParseRole(role.String())
		if err != nil {
			t.Fatalf("ParseRole(%q) error: %v", role, err)
		}
		if got != role {
			t.Fatalf("ParseRole(%q) = %q", role, got)
		}
	}
}

func TestParseRoleInvalid(t *testing.T) {
	_, err := identity.ParseRole("superadmin")
	if err == nil {
		t.Fatal("expected error for unknown role")
	}
}
