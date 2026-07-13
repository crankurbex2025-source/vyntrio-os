package handlers_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func loginAndGetSessionCookie(t *testing.T, router identityRouter) *http.Cookie {
	t.Helper()

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("login status = %d", rec.Code)
	}
	sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie")
	}
	return sessionCookie
}

func TestLogoutRevokesSessionClearsCookiesAndAudits(t *testing.T) {
	secure := true
	router := newIdentityRouter(t, "production", &secure)
	bootstrapOwner(t, router)
	sessionCookie := loginAndGetSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}

	cleared := rec.Result().Cookies()
	if len(cleared) != 1 {
		t.Fatalf("clear cookie count = %d, want 1", len(cleared))
	}
	c := cleared[0]
	if c.Name != cookie.SessionCookieName {
		t.Fatalf("cleared cookie name = %q", c.Name)
	}
	if c.MaxAge != -1 {
		t.Fatalf("cookie MaxAge = %d, want -1", c.MaxAge)
	}
	if c.Value != "" {
		t.Fatalf("cookie value must be empty")
	}
	if c.Path != "/" {
		t.Fatalf("cookie path = %q", c.Path)
	}
	if !c.Secure {
		t.Fatal("cookie must remain Secure in production")
	}
	if c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie SameSite = %v", c.SameSite)
	}
	if !c.HttpOnly {
		t.Fatal("cleared session cookie must stay HttpOnly")
	}

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := router.sessions.GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.RevokedAt == "" {
		t.Fatal("session was not revoked")
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	logoutAudits := 0
	for _, event := range events {
		if event.EventType == "identity.logout.succeeded" {
			logoutAudits++
		}
		if strings.Contains(event.MetadataJSON, sessionCookie.Value) {
			t.Fatalf("audit leaked cookie: %q", event.MetadataJSON)
		}
	}
	if logoutAudits != 1 {
		t.Fatalf("logout audit count = %d, want 1", logoutAudits)
	}
}

func TestLogoutIsIdempotent(t *testing.T) {
	router := newIdentityRouter(t, "development", nil)
	bootstrapOwner(t, router)
	sessionCookie := loginAndGetSessionCookie(t, router)

	rec1 := httptest.NewRecorder()
	router.handler.ServeHTTP(rec1, logoutPOST([]*http.Cookie{sessionCookie}))
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("first logout status = %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	router.handler.ServeHTTP(rec2, logoutPOST([]*http.Cookie{sessionCookie}))
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("second logout status = %d", rec2.Code)
	}
	if len(rec2.Result().Cookies()) != 1 {
		t.Fatalf("second logout must still clear session cookie")
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	logoutAudits := 0
	for _, event := range events {
		if event.EventType == "identity.logout.succeeded" {
			logoutAudits++
		}
	}
	if logoutAudits != 1 {
		t.Fatalf("logout audit count = %d, want 1", logoutAudits)
	}
}

func TestLogoutMissingOrInvalidCookieStillSucceeds(t *testing.T) {
	router := newIdentityRouter(t, "development", nil)
	bootstrapOwner(t, router)

	cases := []struct {
		name    string
		cookies []*http.Cookie
	}{
		{name: "missing cookie", cookies: nil},
		{name: "unknown token", cookies: []*http.Cookie{{Name: cookie.SessionCookieName, Value: "unknown-token-value"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, logoutPOST(tc.cookies))
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want 204", rec.Code)
			}
			if len(rec.Result().Cookies()) != 1 {
				t.Fatalf("must clear session cookie")
			}
		})
	}
}

func TestLogoutExpiredSessionDoesNotAudit(t *testing.T) {
	router := newIdentityRouter(t, "development", nil)
	bootstrapOwner(t, router)
	sessionCookie := loginAndGetSessionCookie(t, router)

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := router.sessions.GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	_, err = router.store.DB().ExecContext(
		context.Background(),
		`UPDATE sessions SET expires_at = ?, idle_expires_at = ? WHERE id = ?`,
		"2000-01-01T00:00:00Z",
		"2000-01-01T00:00:00Z",
		cred.Session.ID,
	)
	if err != nil {
		t.Fatalf("expire session: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range events {
		if event.EventType == "identity.logout.succeeded" {
			t.Fatal("expired logout must not write logout audit")
		}
	}
}

func TestLogoutDoesNotCreateSession(t *testing.T) {
	router := newIdentityRouter(t, "development", nil)
	bootstrapOwner(t, router)

	countBefore, err := countSessions(context.Background(), router.store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST(nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}

	countAfter, err := countSessions(context.Background(), router.store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}
	if countBefore != countAfter {
		t.Fatalf("session count changed from %d to %d", countBefore, countAfter)
	}
}

type failingLogoutRevoker struct{}

func (f *failingLogoutRevoker) RevokeActiveSessionByTokenHash(
	ctx context.Context,
	sessionTokenHash, revokedAt string,
	audit appidentity.AppendSecurityAuditEventInput,
) (bool, error) {
	return false, errors.New("revoke failed")
}

func TestLogoutRevokeFailureStillClearsCookies(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	hasher, err := appidentity.NewPasswordHasher(appidentity.Argon2idConfig{
		Memory: 4096, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() error: %v", err)
	}
	userRepo := sqlite.NewUserRepository(store.DB())
	bootstrapRepo := sqlite.NewBootstrapRepository(store.DB())
	bootstrapService := appidentity.NewBootstrapService(hasher, bootstrapRepo, userRepo)
	bootstrap := handlers.NewBootstrap(handlers.BootstrapDeps{Service: bootstrapService})

	sessionTokens, err := appidentity.NewSessionTokenService(appidentity.DefaultSessionTokenConfig)
	if err != nil {
		t.Fatalf("NewSessionTokenService() error: %v", err)
	}
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, sqlite.NewLoginRepository(store.DB()))
	login := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookie.NewPolicy("development", nil),
	})
	logoutService := appidentity.NewLogoutService(&failingLogoutRevoker{})
	logout := handlers.NewLogout(handlers.LogoutDeps{
		Service:      logoutService,
		CookiePolicy: cookie.NewPolicy("development", nil),
	})

	router := identityRouter{
		handler: httpapiNewRouter(store, bootstrap, login, logout),
		store:   store,
	}

	bootstrapOwner(t, router)
	sessionCookie := loginAndGetSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if len(rec.Result().Cookies()) != 1 {
		t.Fatalf("must still clear session cookie on revoke failure")
	}

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := sqlite.NewSessionRepository(store.DB()).GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.RevokedAt != "" {
		t.Fatal("session must not be marked revoked when revoke failed")
	}
}

func httpapiNewRouter(store *sqlite.Store, bootstrap *handlers.Bootstrap, login *handlers.Login, logout *handlers.Logout) http.Handler {
	cfg := config.Config{Env: "development", ReadTimeout: 15 * time.Second}
	return httpapi.NewRouter(cfg, slog.Default(), health.NewReadiness(store), bootstrap, login, logout, nil, nil)
}
