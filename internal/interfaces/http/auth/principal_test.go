package auth_test

import (
	"context"
	"testing"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/auth"
)

func TestPrincipalFromContextReturnsStoredPrincipal(t *testing.T) {
	want := auth.Principal{
		UserID: domainidentity.UserID("user-1"),
		Role:   domainidentity.RoleOwner,
	}
	ctx := auth.WithPrincipal(context.Background(), want)

	got, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		t.Fatal("PrincipalFromContext() = false")
	}
	if got != want {
		t.Fatalf("principal = %+v, want %+v", got, want)
	}
}

func TestPrincipalFromContextMissingFailsClosed(t *testing.T) {
	if _, ok := auth.PrincipalFromContext(context.Background()); ok {
		t.Fatal("expected missing principal to fail closed")
	}
}

func TestPrincipalFromContextInvalidRoleFailsClosed(t *testing.T) {
	ctx := auth.WithPrincipal(context.Background(), auth.Principal{
		UserID: domainidentity.UserID("user-1"),
		Role:   domainidentity.Role("bad"),
	})
	if _, ok := auth.PrincipalFromContext(ctx); ok {
		t.Fatal("expected invalid role to fail closed")
	}
}

func TestPrincipalDomainPrincipal(t *testing.T) {
	p := auth.Principal{
		UserID: domainidentity.UserID("user-1"),
		Role:   domainidentity.RoleOperator,
	}
	domainPrincipal := p.DomainPrincipal()
	if !domainPrincipal.Authenticated || domainPrincipal.UserID != p.UserID || domainPrincipal.Role != p.Role {
		t.Fatalf("domain principal = %+v", domainPrincipal)
	}
}
