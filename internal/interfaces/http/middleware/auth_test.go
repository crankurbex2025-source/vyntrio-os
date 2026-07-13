package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/auth"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
)

const (
	testRawSessionToken = "probe-session-token-value"
	futureSessionExpiry = "2099-01-01T00:00:00Z"
)

type authTestEnv struct {
	store    *sqlite.Store
	resolver *appidentity.SessionResolver
	authz    ports.Authorizer
}

func newAuthTestEnv(t *testing.T) authTestEnv {
	t.Helper()

	store, err := sqlite.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	return authTestEnv{
		store:    store,
		resolver: appidentity.NewSessionResolver(sqlite.NewSessionAuthRepository(store.DB())),
		authz:    ports.NewRBACAuthorizer(),
	}
}

func (env authTestEnv) createActiveOwnerSession(t *testing.T, rawToken, csrfHash string) {
	t.Helper()
	ctx := context.Background()
	users := sqlite.NewUserRepository(env.store.DB())
	sessions := sqlite.NewSessionRepository(env.store.DB())

	if err := users.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           domainidentity.UserID("owner-1"),
		Username:     "owner",
		PasswordHash: "hash-owner",
		Role:         domainidentity.RoleOwner,
		Status:       appidentity.UserStatusActive,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-owner",
		UserID:           domainidentity.UserID("owner-1"),
		SessionTokenHash: appidentity.HashRawToken(rawToken),
		CSRFTokenHash:    csrfHash,
		CreatedAt:        "2026-07-13T10:00:00Z",
		LastSeenAt:       "2026-07-13T10:00:00Z",
		ExpiresAt:        futureSessionExpiry,
		IdleExpiresAt:    futureSessionExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}
}

func (env authTestEnv) createReadOnlySession(t *testing.T, rawToken string) {
	t.Helper()
	ctx := context.Background()
	users := sqlite.NewUserRepository(env.store.DB())
	sessions := sqlite.NewSessionRepository(env.store.DB())

	if err := users.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           domainidentity.UserID("reader-1"),
		Username:     "reader",
		PasswordHash: "hash-reader",
		Role:         domainidentity.RoleReadOnly,
		Status:       appidentity.UserStatusActive,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	if err := sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "sess-reader",
		UserID:           domainidentity.UserID("reader-1"),
		SessionTokenHash: appidentity.HashRawToken(rawToken),
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        "2026-07-13T10:00:00Z",
		LastSeenAt:       "2026-07-13T10:00:00Z",
		ExpiresAt:        futureSessionExpiry,
		IdleExpiresAt:    futureSessionExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}
}

func serveWithRequestID(req *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	wrapped := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}))
	wrapped.ServeHTTP(rec, req)
	return rec
}

func protectedProbeHandler(t *testing.T, env authTestEnv, perm domainidentity.Permission, onHit func(r *http.Request)) http.Handler {
	t.Helper()
	return middleware.OptionalAuthentication(env.resolver)(
		middleware.RequireAuthentication(
			middleware.RequirePermission(env.authz, perm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if onHit != nil {
					onHit(r)
				}
				w.WriteHeader(http.StatusNoContent)
			})),
		),
	)
}

func parseErrorCode(t *testing.T, body []byte) string {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error object")
	}
	code, _ := errObj["code"].(string)
	return code
}

func TestOptionalAuthenticationValidSessionSetsPrincipal(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	var gotPrincipal auth.Principal
	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("missing principal")
		}
		gotPrincipal = p
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if gotPrincipal.UserID != domainidentity.UserID("owner-1") || gotPrincipal.Role != domainidentity.RoleOwner {
		t.Fatalf("principal = %+v", gotPrincipal)
	}
}

func TestOptionalAuthenticationMissingCookieIsAnonymous(t *testing.T) {
	env := newAuthTestEnv(t)
	called := false
	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := auth.PrincipalFromContext(r.Context()); ok {
			t.Fatal("expected anonymous request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := serveWithRequestID(httptest.NewRequest(http.MethodGet, "/probe", nil), handler)
	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("status = %d called = %v", rec.Code, called)
	}
}

func TestOptionalAuthenticationOversizedCookieIsAnonymous(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	oversized := strings.Repeat("a", appidentity.MaxSessionCookieValueLen+1)
	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.PrincipalFromContext(r.Context()); ok {
			t.Fatal("expected anonymous request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: oversized})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequireAuthenticationAnonymousReturns401(t *testing.T) {
	handler := middleware.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not run")
	}))

	rec := serveWithRequestID(httptest.NewRequest(http.MethodGet, "/probe", nil), handler)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if parseErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestRequireAuthenticationAuthenticatedReachesNext(t *testing.T) {
	called := false
	handler := middleware.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID: domainidentity.UserID("owner-1"),
		Role:   domainidentity.RoleOwner,
	}))
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("status = %d called = %v", rec.Code, called)
	}
}

func TestProtectedProbeOwnerAllowed(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	handler := protectedProbeHandler(t, env, domainidentity.PermissionRolesAssignOwner, nil)
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestProtectedProbeReadOnlyDenied(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createReadOnlySession(t, "reader-session-token")

	handler := protectedProbeHandler(t, env, domainidentity.PermissionRolesAssignOwner, nil)
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: "reader-session-token"})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if parseErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestProtectedProbeAnonymousReturns401(t *testing.T) {
	env := newAuthTestEnv(t)
	handler := protectedProbeHandler(t, env, domainidentity.PermissionRolesAssignOwner, func(r *http.Request) {
		t.Fatal("handler must not run")
	})

	rec := serveWithRequestID(httptest.NewRequest(http.MethodGet, "/probe", nil), handler)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestProtectedProbeInvalidPermissionDenied(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	handler := protectedProbeHandler(t, env, domainidentity.Permission("secrets:dump"), nil)
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequireAuthenticationIndistinguishable401Cases(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	cases := []struct {
		name string
		req  *http.Request
	}{
		{name: "missing cookie", req: httptest.NewRequest(http.MethodGet, "/probe", nil)},
		{name: "unknown token", req: func() *http.Request {
			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: "unknown-token"})
			return req
		}()},
	}

	var referenceBody []byte
	for i, tc := range cases {
		handler := middleware.OptionalAuthentication(env.resolver)(
			middleware.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("next handler must not run")
			})),
		)
		rec := serveWithRequestID(tc.req, handler)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s status = %d", tc.name, rec.Code)
		}
		body := normalizeAuthErrorBody(rec.Body.Bytes())
		if i == 0 {
			referenceBody = body
		} else if string(body) != string(referenceBody) {
			t.Fatalf("%s body differs: %s", tc.name, rec.Body.String())
		}
	}
}

func normalizeAuthErrorBody(body []byte) []byte {
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	if errObj, ok := payload["error"].(map[string]any); ok {
		delete(errObj, "request_id")
	}
	out, _ := json.Marshal(payload)
	return out
}

type countingSessionAuthStore struct {
	inner appidentity.SessionAuthStore
	calls atomic.Int32
}

func (s *countingSessionAuthStore) GetSessionAuthByTokenHash(
	ctx context.Context,
	tokenHash string,
) (appidentity.SessionAuthRecord, error) {
	s.calls.Add(1)
	return s.inner.GetSessionAuthByTokenHash(ctx, tokenHash)
}

func TestAuthMiddlewareCompositionResolvesOnce(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	counter := &countingSessionAuthStore{inner: sqlite.NewSessionAuthRepository(env.store.DB())}
	resolver := appidentity.NewSessionResolver(counter)

	handler := middleware.OptionalAuthentication(resolver)(
		middleware.RequireAuthentication(
			middleware.RequirePermission(env.authz, domainidentity.PermissionRolesAssignOwner)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				}),
			),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if counter.calls.Load() != 1 {
		t.Fatalf("lookup calls = %d, want 1", counter.calls.Load())
	}
}

type failingSessionAuthStore struct{}

func (failingSessionAuthStore) GetSessionAuthByTokenHash(
	context.Context,
	string,
) (appidentity.SessionAuthRecord, error) {
	return appidentity.SessionAuthRecord{}, errors.New("database unavailable")
}

func TestOptionalAuthenticationStoreFailureReturns500(t *testing.T) {
	resolver := appidentity.NewSessionResolver(failingSessionAuthStore{})
	handler := middleware.OptionalAuthentication(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not run")
	}))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: "token"})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSessionResolverUsesSameHashAsLoginPersistence(t *testing.T) {
	env := newAuthTestEnv(t)
	raw := "hash-consistency-token"
	env.createActiveOwnerSession(t, raw, appidentity.HashRawToken("auth-test-csrf"))

	record, err := sqlite.NewSessionAuthRepository(env.store.DB()).GetSessionAuthByTokenHash(
		context.Background(),
		appidentity.HashRawToken(raw),
	)
	if err != nil {
		t.Fatalf("GetSessionAuthByTokenHash() error: %v", err)
	}
	if record.UserID != domainidentity.UserID("owner-1") {
		t.Fatalf("user_id = %q", record.UserID)
	}
}

func TestOptionalAuthenticationRevokedSessionIsAnonymous(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	_, err := env.store.DB().ExecContext(
		context.Background(),
		`UPDATE sessions SET revoked_at = ? WHERE id = ?`,
		"2026-07-13T12:00:00Z",
		"sess-owner",
	)
	if err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.PrincipalFromContext(r.Context()); ok {
			t.Fatal("expected anonymous request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOptionalAuthenticationIdleExpiredSessionIsAnonymous(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	_, err := env.store.DB().ExecContext(
		context.Background(),
		`UPDATE sessions SET idle_expires_at = ? WHERE id = ?`,
		"2000-01-01T00:00:00Z",
		"sess-owner",
	)
	if err != nil {
		t.Fatalf("expire idle session: %v", err)
	}

	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.PrincipalFromContext(r.Context()); ok {
			t.Fatal("expected anonymous request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOptionalAuthenticationCancellationReturns408(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createActiveOwnerSession(t, testRawSessionToken, appidentity.HashRawToken("auth-test-csrf"))

	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not run")
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil).WithContext(ctx)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: testRawSessionToken})
	rec := serveWithRequestID(req, handler)
	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOptionalAuthenticationDoesNotModifyCookies(t *testing.T) {
	env := newAuthTestEnv(t)
	handler := middleware.OptionalAuthentication(env.resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := serveWithRequestID(httptest.NewRequest(http.MethodGet, "/probe", nil), handler)
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("Set-Cookie count = %d, want 0", len(rec.Result().Cookies()))
	}
}

func TestProtectedProbeHandlerNotCalledAfter401(t *testing.T) {
	env := newAuthTestEnv(t)
	called := false
	handler := protectedProbeHandler(t, env, domainidentity.PermissionRolesAssignOwner, func(r *http.Request) {
		called = true
	})
	serveWithRequestID(httptest.NewRequest(http.MethodGet, "/probe", nil), handler)
	if called {
		t.Fatal("handler executed for anonymous request")
	}
}

func TestProtectedProbeHandlerNotCalledAfter403(t *testing.T) {
	env := newAuthTestEnv(t)
	env.createReadOnlySession(t, "reader-session-token")
	called := false
	handler := protectedProbeHandler(t, env, domainidentity.PermissionRolesAssignOwner, func(r *http.Request) {
		called = true
	})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: "reader-session-token"})
	serveWithRequestID(req, handler)
	if called {
		t.Fatal("handler executed after forbidden")
	}
}
