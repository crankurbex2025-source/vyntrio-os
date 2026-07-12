package ports

import "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"

// Authorizer evaluates whether a principal may exercise a permission.
type Authorizer interface {
	Authorize(principal identity.Principal, perm identity.Permission) error
}

// RBACAuthorizer enforces ADR-0004 role permissions for authenticated principals.
type RBACAuthorizer struct{}

// NewRBACAuthorizer creates a default role-based authorizer.
func NewRBACAuthorizer() *RBACAuthorizer {
	return &RBACAuthorizer{}
}

// Authorize returns nil when principal is authenticated and role grants permission.
func (RBACAuthorizer) Authorize(principal identity.Principal, perm identity.Permission) error {
	if !principal.Authenticated {
		return identity.ErrUnauthorized
	}
	if !principal.Role.Valid() {
		return identity.ErrForbidden
	}
	if !perm.Valid() {
		return identity.ErrForbidden
	}
	if !identity.Allows(principal.Role, perm) {
		return identity.ErrForbidden
	}
	return nil
}
