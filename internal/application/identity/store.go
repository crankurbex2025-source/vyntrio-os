// Package identity defines application ports for identity persistence.
package identity

import (
	"context"
	"errors"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

// ErrNotFound indicates the requested identity record does not exist.
var ErrNotFound = errors.New("identity record not found")

// UserStatus is the persisted user lifecycle state.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// User is a safe read model without credential material.
type User struct {
	ID                 domainidentity.UserID
	Username           string
	DisplayName        string
	Role               domainidentity.Role
	Status             UserStatus
	MustChangePassword bool
	CreatedAt          string
	UpdatedAt          string
	LastLoginAt        string
}

// UserCredential includes password hash for credential lookup only.
type UserCredential struct {
	User         User
	PasswordHash string
}

// CreateUserInput holds data required to create a user.
type CreateUserInput struct {
	ID                 domainidentity.UserID
	Username           string
	DisplayName        string
	PasswordHash       string
	Role               domainidentity.Role
	Status             UserStatus
	MustChangePassword bool
}

// ListUsersInput bounds deterministic user listing ordered by username.
type ListUsersInput struct {
	Limit         int
	AfterUsername string
}

// Session is a safe read model without token hash material.
type Session struct {
	ID            string
	UserID        domainidentity.UserID
	CreatedAt     string
	ExpiresAt     string
	IdleExpiresAt string
	LastSeenAt    string
	RevokedAt     string
}

// SessionCredential includes hashed token material for internal validation lookup only.
type SessionCredential struct {
	Session          Session
	SessionTokenHash string
	CSRFTokenHash    string
	UserAgentHash    string
	IPHash           string
}

// CreateSessionInput holds data required to create a session.
type CreateSessionInput struct {
	ID               string
	UserID           domainidentity.UserID
	SessionTokenHash string
	CSRFTokenHash    string
	ExpiresAt        string
	IdleExpiresAt    string
	UserAgentHash    string
	IPHash           string
}

// SecurityAuditEvent is an append-only audit record.
type SecurityAuditEvent struct {
	ID            string
	OccurredAt    string
	ActorUserID   domainidentity.UserID
	SubjectUserID domainidentity.UserID
	EventType     string
	Result        string
	IPHash        string
	UserAgentHash string
	MetadataJSON  string
}

// AppendSecurityAuditEventInput holds data required to append an audit event.
type AppendSecurityAuditEventInput struct {
	ID            string
	ActorUserID   domainidentity.UserID
	SubjectUserID domainidentity.UserID
	EventType     string
	Result        string
	IPHash        string
	UserAgentHash string
	MetadataJSON  string
}

// AuditListCursor identifies the last item from a prior audit page.
type AuditListCursor struct {
	OccurredAt string
	ID         string
}

// ListSecurityAuditEventsInput bounds cursor pagination ordered newest first.
type ListSecurityAuditEventsInput struct {
	Limit int
	After *AuditListCursor
}

// UserStore persists user records.
type UserStore interface {
	CreateUser(ctx context.Context, input CreateUserInput) error
	GetUserByID(ctx context.Context, id domainidentity.UserID) (User, error)
	GetUserByUsername(ctx context.Context, username string) (UserCredential, error)
	UpdateUserPasswordHash(ctx context.Context, id domainidentity.UserID, passwordHash string) error
	UpdateUserLastLoginAt(ctx context.Context, id domainidentity.UserID, lastLoginAt string) error
	SetUserStatus(ctx context.Context, id domainidentity.UserID, status UserStatus) error
	ListUsers(ctx context.Context, input ListUsersInput) ([]User, error)
	CountUsers(ctx context.Context) (int64, error)
}

// SessionStore persists session records.
type SessionStore interface {
	CreateSession(ctx context.Context, input CreateSessionInput) error
	GetSessionByTokenHash(ctx context.Context, sessionTokenHash string) (SessionCredential, error)
	TouchSession(ctx context.Context, id string, lastSeenAt string) error
	RevokeSessionByID(ctx context.Context, id string, revokedAt string) error
	RevokeAllSessionsForUser(ctx context.Context, userID domainidentity.UserID, revokedAt string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
}

// SecurityAuditStore persists append-only security audit events.
type SecurityAuditStore interface {
	AppendSecurityAuditEvent(ctx context.Context, input AppendSecurityAuditEventInput) error
	ListSecurityAuditEvents(ctx context.Context, input ListSecurityAuditEventsInput) ([]SecurityAuditEvent, error)
}
