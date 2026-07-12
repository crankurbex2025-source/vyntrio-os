package identity

import "errors"

var (
	// ErrInvalidRole indicates the role is not one of the five fixed v1 roles.
	ErrInvalidRole = errors.New("invalid role")
	// ErrInvalidPermission indicates the permission is not normative v1.
	ErrInvalidPermission = errors.New("invalid permission")
	// ErrUnauthorized indicates the principal is not authenticated.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden indicates the principal lacks the requested permission.
	ErrForbidden = errors.New("forbidden")
)
