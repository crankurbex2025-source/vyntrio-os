package sqlite_test

import (
	"errors"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestSessionAuthRepositoryLookupByTokenHash(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())
	authRepo := sqlite.NewSessionAuthRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")
	if err := users.SetUserStatus(ctx, domainidentity.UserID("user-1"), appidentity.UserStatusActive); err != nil {
		t.Fatalf("SetUserStatus() error: %v", err)
	}

	rawToken := "integration-session-token"
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-1",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: appidentity.HashRawToken(rawToken),
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        sessionCreatedAt,
		LastSeenAt:       sessionCreatedAt,
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	record, err := authRepo.GetSessionAuthByTokenHash(ctx, appidentity.HashRawToken(rawToken))
	if err != nil {
		t.Fatalf("GetSessionAuthByTokenHash() error: %v", err)
	}
	if record.SessionID != "sess-1" || record.UserID != domainidentity.UserID("user-1") {
		t.Fatalf("record = %+v", record)
	}
	if record.Role != domainidentity.RoleAdministrator {
		t.Fatalf("role = %q", record.Role)
	}
	if record.UserStatus != appidentity.UserStatusActive {
		t.Fatalf("status = %q", record.UserStatus)
	}
}

func TestSessionAuthRepositoryMissingTokenNotFound(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	authRepo := sqlite.NewSessionAuthRepository(store.DB())

	_, err := authRepo.GetSessionAuthByTokenHash(ctx, "missing-hash")
	if !errors.Is(err, appidentity.ErrNotFound) {
		t.Fatalf("GetSessionAuthByTokenHash() error = %v, want ErrNotFound", err)
	}
}

func TestSessionAuthRepositoryJoinsDisabledUser(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())
	authRepo := sqlite.NewSessionAuthRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")
	if err := users.SetUserStatus(ctx, domainidentity.UserID("user-1"), appidentity.UserStatusDisabled); err != nil {
		t.Fatalf("SetUserStatus() error: %v", err)
	}

	rawToken := "disabled-user-token"
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-disabled",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: appidentity.HashRawToken(rawToken),
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        sessionCreatedAt,
		LastSeenAt:       sessionCreatedAt,
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	record, err := authRepo.GetSessionAuthByTokenHash(ctx, appidentity.HashRawToken(rawToken))
	if err != nil {
		t.Fatalf("GetSessionAuthByTokenHash() error: %v", err)
	}
	if record.UserStatus != appidentity.UserStatusDisabled {
		t.Fatalf("status = %q", record.UserStatus)
	}
}
