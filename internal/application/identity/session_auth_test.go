package identity_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

type stubSessionAuthStore struct {
	record appidentity.SessionAuthRecord
	err    error
	calls  int
}

func (s *stubSessionAuthStore) GetSessionAuthByTokenHash(
	_ context.Context,
	_ string,
) (appidentity.SessionAuthRecord, error) {
	s.calls++
	if s.err != nil {
		return appidentity.SessionAuthRecord{}, s.err
	}
	return s.record, nil
}

func validSessionAuthRecord() appidentity.SessionAuthRecord {
	return appidentity.SessionAuthRecord{
		SessionID:     "sess-1",
		UserID:        domainidentity.UserID("user-1"),
		CSRFTokenHash: "csrf-hash-value",
		ExpiresAt:     "2099-01-01T00:00:00Z",
		IdleExpiresAt: "2099-01-01T00:00:00Z",
		UserStatus:    appidentity.UserStatusActive,
		Role:          domainidentity.RoleOwner,
	}
}

func TestSessionResolverMissingCookieIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v, want anonymous", ok, err)
	}
	if store.calls != 0 {
		t.Fatalf("lookup calls = %d, want 0", store.calls)
	}
}

func TestSessionResolverOversizedCookieIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{}
	resolver := appidentity.NewSessionResolver(store)

	raw := make([]byte, appidentity.MaxSessionCookieValueLen+1)
	for i := range raw {
		raw[i] = 'a'
	}

	_, ok, err := resolver.Resolve(context.Background(), string(raw))
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v, want anonymous", ok, err)
	}
	if store.calls != 0 {
		t.Fatalf("lookup calls = %d, want 0", store.calls)
	}
}

func TestSessionResolverValidSessionReturnsSubject(t *testing.T) {
	fixed := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	store := &stubSessionAuthStore{record: validSessionAuthRecord()}
	resolver := appidentity.NewSessionResolver(store)

	session, ok, err := resolver.Resolve(context.Background(), "valid-session-token-value")
	if err != nil || !ok {
		t.Fatalf("Resolve() = (%+v,%v,%v)", session, ok, err)
	}
	if session.UserID != domainidentity.UserID("user-1") || session.Role != domainidentity.RoleOwner {
		t.Fatalf("subject = (%q,%q)", session.UserID, session.Role)
	}
	if session.CSRFTokenHash != "csrf-hash-value" {
		t.Fatalf("csrf hash = %q", session.CSRFTokenHash)
	}
	if store.calls != 1 {
		t.Fatalf("lookup calls = %d", store.calls)
	}
	_ = fixed
}

func TestSessionResolverUnknownTokenIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{err: appidentity.ErrNotFound}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "missing-token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverStoreFailureReturnsError(t *testing.T) {
	store := &stubSessionAuthStore{err: errors.New("db unavailable")}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "token")
	if err == nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v, want error", ok, err)
	}
}

func TestSessionResolverCancellationReturnsError(t *testing.T) {
	store := &stubSessionAuthStore{}
	resolver := appidentity.NewSessionResolver(store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, ok, err := resolver.Resolve(ctx, "token")
	if !errors.Is(err, context.Canceled) || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
	if store.calls != 0 {
		t.Fatalf("lookup calls = %d, want 0", store.calls)
	}
}

func TestSessionResolverRevokedSessionIsAnonymous(t *testing.T) {
	record := validSessionAuthRecord()
	record.RevokedAt = "2026-07-13T00:00:00Z"
	store := &stubSessionAuthStore{record: record}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverDisabledUserIsAnonymous(t *testing.T) {
	record := validSessionAuthRecord()
	record.UserStatus = appidentity.UserStatusDisabled
	store := &stubSessionAuthStore{record: record}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverInvalidRoleIsAnonymous(t *testing.T) {
	record := validSessionAuthRecord()
	record.Role = domainidentity.Role("bad")
	store := &stubSessionAuthStore{record: record}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverExpiredSessionIsAnonymous(t *testing.T) {
	record := validSessionAuthRecord()
	record.ExpiresAt = "2000-01-01T00:00:00Z"
	record.IdleExpiresAt = "2000-01-01T00:00:00Z"
	store := &stubSessionAuthStore{record: record}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverMissingCSRFHashIsAnonymous(t *testing.T) {
	record := validSessionAuthRecord()
	record.CSRFTokenHash = ""
	store := &stubSessionAuthStore{record: record}
	resolver := appidentity.NewSessionResolver(store)

	_, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}
