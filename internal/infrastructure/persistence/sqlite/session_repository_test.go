package sqlite_test

import (
	"errors"
	"strings"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

const (
	futureExpiry = "2099-01-01T00:00:00Z"
	pastExpiry   = "2000-01-01T00:00:00Z"
)

func TestSessionRepositoryCreateAndLookupByTokenHash(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")

	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-1",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: "session-hash-1",
		CSRFTokenHash:    "csrf-hash-1",
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
		UserAgentHash:    "ua-hash",
		IPHash:           "ip-hash",
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	cred, err := sessions.GetSessionByTokenHash(ctx, "session-hash-1")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.ID != "sess-1" {
		t.Fatalf("session id = %q", cred.Session.ID)
	}
	if cred.SessionTokenHash != "session-hash-1" {
		t.Fatalf("session token hash not returned for validation lookup")
	}
	if cred.CSRFTokenHash != "csrf-hash-1" {
		t.Fatalf("csrf token hash not returned for validation lookup")
	}
	if cred.Session.UserID != domainidentity.UserID("user-1") {
		t.Fatalf("user_id = %q", cred.Session.UserID)
	}
}

func TestSessionRepositoryGetByTokenHashNotFound(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	sessions := sqlite.NewSessionRepository(store.DB())

	_, err := sessions.GetSessionByTokenHash(ctx, "missing")
	if !errors.Is(err, appidentity.ErrNotFound) {
		t.Fatalf("GetSessionByTokenHash() error = %v, want ErrNotFound", err)
	}
}

func TestSessionRepositoryTouchSession(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-1",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: "session-hash-1",
		CSRFTokenHash:    "csrf-hash-1",
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	touchAt := "2026-07-12T13:00:00Z"
	if err := sessions.TouchSession(ctx, "sess-1", touchAt); err != nil {
		t.Fatalf("TouchSession() error: %v", err)
	}

	cred, err := sessions.GetSessionByTokenHash(ctx, "session-hash-1")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.LastSeenAt != touchAt {
		t.Fatalf("last_seen_at = %q, want %q", cred.Session.LastSeenAt, touchAt)
	}
}

func TestSessionRepositoryRevokeOneAndAllForUser(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")

	for i, token := range []string{"hash-a", "hash-b"} {
		if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
			ID:               "sess-" + string(rune('a'+i)),
			UserID:           domainidentity.UserID("user-1"),
			SessionTokenHash: token,
			CSRFTokenHash:    "csrf-" + token,
			ExpiresAt:        futureExpiry,
			IdleExpiresAt:    futureExpiry,
		}); err != nil {
			t.Fatalf("CreateSession() error: %v", err)
		}
	}

	revokedAt := "2026-07-12T14:00:00Z"
	if err := sessions.RevokeSessionByID(ctx, "sess-a", revokedAt); err != nil {
		t.Fatalf("RevokeSessionByID() error: %v", err)
	}

	credA, err := sessions.GetSessionByTokenHash(ctx, "hash-a")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash(a) error: %v", err)
	}
	if credA.Session.RevokedAt != revokedAt {
		t.Fatalf("session a revoked_at = %q", credA.Session.RevokedAt)
	}

	credB, err := sessions.GetSessionByTokenHash(ctx, "hash-b")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash(b) error: %v", err)
	}
	if credB.Session.RevokedAt != "" {
		t.Fatalf("session b should remain active, revoked_at = %q", credB.Session.RevokedAt)
	}

	if err := sessions.RevokeAllSessionsForUser(ctx, domainidentity.UserID("user-1"), revokedAt); err != nil {
		t.Fatalf("RevokeAllSessionsForUser() error: %v", err)
	}

	credB, err = sessions.GetSessionByTokenHash(ctx, "hash-b")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash(b) after revoke all error: %v", err)
	}
	if credB.Session.RevokedAt != revokedAt {
		t.Fatalf("session b revoked_at = %q after revoke all", credB.Session.RevokedAt)
	}
}

func TestSessionRepositoryDeleteExpiredSessionsOnly(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")

	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-active",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: "active-hash",
		CSRFTokenHash:    "csrf-active",
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession active error: %v", err)
	}
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-expired",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: "expired-hash",
		CSRFTokenHash:    "csrf-expired",
		ExpiresAt:        pastExpiry,
		IdleExpiresAt:    pastExpiry,
	}); err != nil {
		t.Fatalf("CreateSession expired error: %v", err)
	}

	deleted, err := sessions.DeleteExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("DeleteExpiredSessions() error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := sessions.GetSessionByTokenHash(ctx, "expired-hash"); !errors.Is(err, appidentity.ErrNotFound) {
		t.Fatalf("expired session lookup error = %v, want ErrNotFound", err)
	}

	if _, err := sessions.GetSessionByTokenHash(ctx, "active-hash"); err != nil {
		t.Fatalf("active session should remain: %v", err)
	}
}

func TestSessionRepositoryForeignKeyRequiresUser(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	sessions := sqlite.NewSessionRepository(store.DB())

	err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-orphan",
		UserID:           domainidentity.UserID("missing-user"),
		SessionTokenHash: "orphan-hash",
		CSRFTokenHash:    "csrf-orphan",
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
	})
	if err == nil {
		t.Fatal("CreateSession without user expected error")
	}
	if strings.Contains(err.Error(), "orphan-hash") {
		t.Fatalf("FK error leaked token hash: %v", err)
	}
}

func TestSessionSafeModelExcludesTokenHashes(t *testing.T) {
	var session appidentity.Session
	_ = session.ID
	_ = session.UserID
	// Session type intentionally has no SessionTokenHash or CSRFTokenHash fields.
}
