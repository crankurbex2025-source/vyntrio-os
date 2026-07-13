package identity

import (
	"context"
	"testing"
	"time"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

type loginTimestampUserStore struct {
	credential UserCredential
}

func (s loginTimestampUserStore) CreateUser(context.Context, CreateUserInput) error {
	panic("unused")
}
func (s loginTimestampUserStore) GetUserByID(context.Context, domainidentity.UserID) (User, error) {
	panic("unused")
}
func (s loginTimestampUserStore) GetUserByUsername(context.Context, string) (UserCredential, error) {
	return s.credential, nil
}
func (s loginTimestampUserStore) UpdateUserPasswordHash(context.Context, domainidentity.UserID, string) error {
	panic("unused")
}
func (s loginTimestampUserStore) UpdateUserLastLoginAt(context.Context, domainidentity.UserID, string) error {
	panic("unused")
}
func (s loginTimestampUserStore) SetUserStatus(context.Context, domainidentity.UserID, UserStatus) error {
	panic("unused")
}
func (s loginTimestampUserStore) ListUsers(context.Context, ListUsersInput) ([]User, error) {
	panic("unused")
}
func (s loginTimestampUserStore) CountUsers(context.Context) (int64, error) {
	panic("unused")
}

type loginTimestampCreator struct {
	input CreateSessionInput
}

func (c *loginTimestampCreator) CreateSessionWithAudit(
	_ context.Context,
	session CreateSessionInput,
	_ AppendSecurityAuditEventInput,
) error {
	c.input = session
	return nil
}

type loginTimestampAuditStore struct{}

func (loginTimestampAuditStore) AppendSecurityAuditEvent(context.Context, AppendSecurityAuditEventInput) error {
	return nil
}

func (loginTimestampAuditStore) ListSecurityAuditEvents(context.Context, ListSecurityAuditEventsInput) ([]SecurityAuditEvent, error) {
	return nil, nil
}

func TestLoginServicePersistsSessionMaterialTimestamps(t *testing.T) {
	fixedNow := time.Date(2026, 7, 13, 10, 15, 30, 0, time.UTC)
	ctx := context.Background()

	hasher, err := NewPasswordHasher(Argon2idConfig{
		Memory: 4096, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() error: %v", err)
	}
	password := "valid-password-123"
	hash, err := hasher.HashPassword(ctx, password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	tokens, err := NewSessionTokenService(DefaultSessionTokenConfig)
	if err != nil {
		t.Fatalf("NewSessionTokenService() error: %v", err)
	}
	material, err := tokens.NewSessionMaterial(ctx, fixedNow)
	if err != nil {
		t.Fatalf("NewSessionMaterial() error: %v", err)
	}

	creator := &loginTimestampCreator{}
	service := NewLoginService(loginTimestampUserStore{
		credential: UserCredential{
			User: User{
				ID:     domainidentity.UserID("user-1"),
				Status: UserStatusActive,
			},
			PasswordHash: hash,
		},
	}, hasher, tokens, creator, loginTimestampAuditStore{})
	service.now = func() time.Time { return fixedNow }

	if _, err := service.Login(ctx, "owner", password, "sess-1", "audit-1"); err != nil {
		t.Fatalf("Login() error: %v", err)
	}

	wantCreated := FormatUTCTime(material.CreatedAt)
	if creator.input.CreatedAt != wantCreated {
		t.Fatalf("created_at = %q, want %q", creator.input.CreatedAt, wantCreated)
	}
	if creator.input.LastSeenAt != wantCreated {
		t.Fatalf("last_seen_at = %q, want %q", creator.input.LastSeenAt, wantCreated)
	}
	if creator.input.ExpiresAt != FormatUTCTime(material.ExpiresAt) {
		t.Fatalf("expires_at = %q, want %q", creator.input.ExpiresAt, FormatUTCTime(material.ExpiresAt))
	}
	if creator.input.IdleExpiresAt != FormatUTCTime(material.IdleExpiresAt) {
		t.Fatalf("idle_expires_at = %q, want %q", creator.input.IdleExpiresAt, FormatUTCTime(material.IdleExpiresAt))
	}
}
