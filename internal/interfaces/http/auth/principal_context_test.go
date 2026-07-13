package auth

import (
	"context"
	"testing"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

func TestPrincipalFromContextWrongTypeFailsClosed(t *testing.T) {
	ctx := context.WithValue(context.Background(), principalContextKey, "not-a-principal")
	if _, ok := PrincipalFromContext(ctx); ok {
		t.Fatal("expected wrong-type context value to fail closed")
	}
}

func TestPrincipalFromContextEmptyUserIDFailsClosed(t *testing.T) {
	ctx := WithPrincipal(context.Background(), Principal{
		UserID: "",
		Role:   domainidentity.RoleOwner,
	})
	if _, ok := PrincipalFromContext(ctx); ok {
		t.Fatal("expected empty user ID to fail closed")
	}
}
