package ports_test

import (
	"errors"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

func TestRBACAuthorizerAllowsOwner(t *testing.T) {
	auth := ports.NewRBACAuthorizer()
	principal := identity.NewPrincipal("user-1", identity.RoleOwner)

	if err := auth.Authorize(principal, identity.PermissionSettingsWrite); err != nil {
		t.Fatalf("Authorize() error: %v", err)
	}
}

func TestRBACAuthorizerRejectsAnonymous(t *testing.T) {
	auth := ports.NewRBACAuthorizer()
	err := auth.Authorize(identity.Anonymous(), identity.PermissionSystemHealth)
	if !errors.Is(err, identity.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestRBACAuthorizerRejectsForbidden(t *testing.T) {
	auth := ports.NewRBACAuthorizer()
	principal := identity.NewPrincipal("user-2", identity.RoleReadOnly)

	err := auth.Authorize(principal, identity.PermissionSettingsWrite)
	if !errors.Is(err, identity.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRBACAuthorizerRejectsUnknownPermission(t *testing.T) {
	auth := ports.NewRBACAuthorizer()
	principal := identity.NewPrincipal("user-3", identity.RoleOwner)

	err := auth.Authorize(principal, identity.Permission("secrets:dump"))
	if !errors.Is(err, identity.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRBACAuthorizerRejectsInvalidRole(t *testing.T) {
	auth := ports.NewRBACAuthorizer()
	principal := identity.Principal{
		UserID:        "user-4",
		Role:          identity.Role("bad"),
		Authenticated: true,
	}

	err := auth.Authorize(principal, identity.PermissionSystemHealth)
	if !errors.Is(err, identity.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
