package sqlite_test

import (
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestLoginRepositoryCreateSessionWithAudit(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	loginRepo := sqlite.NewLoginRepository(store.DB())
	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")

	createdAt := "2026-07-13T10:15:30Z"
	expiresAt := "2026-07-20T10:15:30Z"
	idleExpiresAt := "2026-07-14T10:15:30Z"

	if err := loginRepo.CreateSessionWithAudit(ctx, appidentity.CreateSessionInput{
		ID:               "sess-1",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: "session-hash",
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        createdAt,
		LastSeenAt:       createdAt,
		ExpiresAt:        expiresAt,
		IdleExpiresAt:    idleExpiresAt,
	}, appidentity.AppendSecurityAuditEventInput{
		ID:            "audit-1",
		ActorUserID:   domainidentity.UserID("user-1"),
		SubjectUserID: domainidentity.UserID("user-1"),
		EventType:     "identity.login.succeeded",
		Result:        "success",
		MetadataJSON:  `{}`,
	}); err != nil {
		t.Fatalf("CreateSessionWithAudit() error: %v", err)
	}

	sessions := sqlite.NewSessionRepository(store.DB())
	cred, err := sessions.GetSessionByTokenHash(ctx, "session-hash")
	if err != nil {
		t.Fatalf("session not persisted: %v", err)
	}
	if cred.Session.CreatedAt != createdAt {
		t.Fatalf("created_at = %q, want %q", cred.Session.CreatedAt, createdAt)
	}
	if cred.Session.LastSeenAt != createdAt {
		t.Fatalf("last_seen_at = %q, want %q", cred.Session.LastSeenAt, createdAt)
	}
	if cred.Session.ExpiresAt != expiresAt {
		t.Fatalf("expires_at = %q, want %q", cred.Session.ExpiresAt, expiresAt)
	}
	if cred.Session.IdleExpiresAt != idleExpiresAt {
		t.Fatalf("idle_expires_at = %q, want %q", cred.Session.IdleExpiresAt, idleExpiresAt)
	}

	events, err := auditRepo.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "identity.login.succeeded" {
		t.Fatalf("audit events = %+v", events)
	}
}

func TestLogoutRepositoryRevokeActiveSessionByTokenHash(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	sessions := sqlite.NewSessionRepository(store.DB())
	logoutRepo := sqlite.NewLogoutRepository(store.DB())
	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())

	createTestUser(t, users, "user-1", "alice")
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-1",
		UserID:           domainidentity.UserID("user-1"),
		SessionTokenHash: "session-hash",
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        sessionCreatedAt,
		LastSeenAt:       sessionCreatedAt,
		ExpiresAt:        futureExpiry,
		IdleExpiresAt:    futureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	revoked, err := logoutRepo.RevokeActiveSessionByTokenHash(ctx, "session-hash", "2026-07-13T00:00:00Z", appidentity.AppendSecurityAuditEventInput{
		ID:           "audit-logout",
		EventType:    "identity.logout.succeeded",
		Result:       "success",
		MetadataJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("RevokeActiveSessionByTokenHash() error: %v", err)
	}
	if !revoked {
		t.Fatal("expected session revoked")
	}

	cred, err := sessions.GetSessionByTokenHash(ctx, "session-hash")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.RevokedAt == "" {
		t.Fatal("revoked_at not set")
	}

	events, err := auditRepo.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "identity.logout.succeeded" {
		t.Fatalf("audit events = %+v", events)
	}

	revokedAgain, err := logoutRepo.RevokeActiveSessionByTokenHash(ctx, "session-hash", "2026-07-13T01:00:00Z", appidentity.AppendSecurityAuditEventInput{
		ID:           "audit-logout-2",
		EventType:    "identity.logout.succeeded",
		Result:       "success",
		MetadataJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("second revoke error: %v", err)
	}
	if revokedAgain {
		t.Fatal("already revoked session must not revoke again")
	}
}
