package auth

import (
	"context"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

type contextKey int

const principalContextKey contextKey = 0

// Principal is the minimal authenticated request subject for HTTP handlers.
type Principal struct {
	UserID domainidentity.UserID
	Role   domainidentity.Role
}

// WithPrincipal stores an authenticated principal in context.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, p)
}

// PrincipalFromContext returns the authenticated principal when present and valid.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	value := ctx.Value(principalContextKey)
	if value == nil {
		return Principal{}, false
	}
	p, ok := value.(Principal)
	if !ok || p.UserID == "" || !p.Role.Valid() {
		return Principal{}, false
	}
	return p, true
}

// DomainPrincipal converts the HTTP principal to the domain authorization model.
func (p Principal) DomainPrincipal() domainidentity.Principal {
	return domainidentity.NewPrincipal(p.UserID, p.Role)
}
