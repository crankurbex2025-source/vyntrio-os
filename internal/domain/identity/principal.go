package identity

// UserID identifies an authenticated user.
type UserID string

// String returns the user identifier.
func (id UserID) String() string {
	return string(id)
}

// Principal is the minimal authenticated subject for authorization decisions.
type Principal struct {
	UserID        UserID
	Role          Role
	Authenticated bool
}

// Anonymous returns an unauthenticated principal.
func Anonymous() Principal {
	return Principal{Authenticated: false}
}

// NewPrincipal returns an authenticated principal.
func NewPrincipal(userID UserID, role Role) Principal {
	return Principal{
		UserID:        userID,
		Role:          role,
		Authenticated: true,
	}
}
