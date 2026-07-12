package identity

import "fmt"

// Role is a fixed v1 system role stored on the user record (ADR-0004).
type Role string

const (
	RoleOwner         Role = "owner"
	RoleAdministrator Role = "administrator"
	RoleOperator      Role = "operator"
	RoleUser          Role = "user"
	RoleReadOnly      Role = "read_only"
)

// AllRoles lists every valid v1 role in stable order.
var AllRoles = []Role{
	RoleOwner,
	RoleAdministrator,
	RoleOperator,
	RoleUser,
	RoleReadOnly,
}

// Valid reports whether role is one of the five fixed v1 roles.
func (r Role) Valid() bool {
	_, ok := roleIndex[r]
	return ok
}

// String returns the canonical role identifier.
func (r Role) String() string {
	return string(r)
}

// ParseRole parses a canonical role identifier.
func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.Valid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidRole, s)
	}
	return r, nil
}

var roleIndex = func() map[Role]struct{} {
	m := make(map[Role]struct{}, len(AllRoles))
	for _, r := range AllRoles {
		m[r] = struct{}{}
	}
	return m
}()
