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

func TestSessionResolverMissingCookieIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "")
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

	_, _, ok, err := resolver.Resolve(context.Background(), string(raw))
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v, want anonymous", ok, err)
	}
	if store.calls != 0 {
		t.Fatalf("lookup calls = %d, want 0", store.calls)
	}
}

func TestSessionResolverValidSessionReturnsSubject(t *testing.T) {
	fixed := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	store := &stubSessionAuthStore{
		record: appidentity.SessionAuthRecord{
			SessionID:     "sess-1",
			UserID:        domainidentity.UserID("user-1"),
			ExpiresAt:     "2026-07-20T12:00:00Z",
			IdleExpiresAt: "2026-07-14T12:00:00Z",
			UserStatus:    appidentity.UserStatusActive,
			Role:          domainidentity.RoleOwner,
		},
	}
	resolver := appidentity.NewSessionResolver(store)

	userID, role, ok, err := resolver.Resolve(context.Background(), "valid-session-token-value")
	if err != nil || !ok {
		t.Fatalf("Resolve() = (%q,%q,%v,%v)", userID, role, ok, err)
	}
	if userID != domainidentity.UserID("user-1") || role != domainidentity.RoleOwner {
		t.Fatalf("subject = (%q,%q)", userID, role)
	}
	if store.calls != 1 {
		t.Fatalf("lookup calls = %d", store.calls)
	}
	_ = fixed
}

func TestSessionResolverUnknownTokenIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{err: appidentity.ErrNotFound}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "missing-token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverStoreFailureReturnsError(t *testing.T) {
	store := &stubSessionAuthStore{err: errors.New("db unavailable")}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "token")
	if err == nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v, want error", ok, err)
	}
}

func TestSessionResolverCancellationReturnsError(t *testing.T) {
	store := &stubSessionAuthStore{}
	resolver := appidentity.NewSessionResolver(store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, ok, err := resolver.Resolve(ctx, "token")
	if !errors.Is(err, context.Canceled) || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
	if store.calls != 0 {
		t.Fatalf("lookup calls = %d, want 0", store.calls)
	}
}

func TestSessionResolverRevokedSessionIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{
		record: appidentity.SessionAuthRecord{
			UserID:        domainidentity.UserID("user-1"),
			ExpiresAt:     "2099-01-01T00:00:00Z",
			IdleExpiresAt: "2099-01-01T00:00:00Z",
			RevokedAt:     "2026-07-13T00:00:00Z",
			UserStatus:    appidentity.UserStatusActive,
			Role:          domainidentity.RoleOwner,
		},
	}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverDisabledUserIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{
		record: appidentity.SessionAuthRecord{
			UserID:        domainidentity.UserID("user-1"),
			ExpiresAt:     "2099-01-01T00:00:00Z",
			IdleExpiresAt: "2099-01-01T00:00:00Z",
			UserStatus:    appidentity.UserStatusDisabled,
			Role:          domainidentity.RoleOwner,
		},
	}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverInvalidRoleIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{
		record: appidentity.SessionAuthRecord{
			UserID:        domainidentity.UserID("user-1"),
			ExpiresAt:     "2099-01-01T00:00:00Z",
			IdleExpiresAt: "2099-01-01T00:00:00Z",
			UserStatus:    appidentity.UserStatusActive,
			Role:          domainidentity.Role("bad"),
		},
	}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}

func TestSessionResolverExpiredSessionIsAnonymous(t *testing.T) {
	store := &stubSessionAuthStore{
		record: appidentity.SessionAuthRecord{
			UserID:        domainidentity.UserID("user-1"),
			ExpiresAt:     "2000-01-01T00:00:00Z",
			IdleExpiresAt: "2000-01-01T00:00:00Z",
			UserStatus:    appidentity.UserStatusActive,
			Role:          domainidentity.RoleOwner,
		},
	}
	resolver := appidentity.NewSessionResolver(store)

	_, _, ok, err := resolver.Resolve(context.Background(), "token")
	if err != nil || ok {
		t.Fatalf("Resolve() = ok=%v err=%v", ok, err)
	}
}
