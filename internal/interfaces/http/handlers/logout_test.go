package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func TestLogoutRevokesSessionClearsCookiesAndAudits(t *testing.T) {
	router := newIdentityRouter(t, true)
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
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
		if event.EventType == appidentity.AuditEventLogoutSucceeded {
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
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec1 := httptest.NewRecorder()
	router.handler.ServeHTTP(rec1, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("first logout status = %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	router.handler.ServeHTTP(rec2, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("second logout status = %d, want 401", rec2.Code)
	}
	if len(rec2.Result().Cookies()) != 0 {
		t.Fatal("second logout must not clear session cookie without valid auth")
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	logoutAudits := 0
	for _, event := range events {
		if event.EventType == appidentity.AuditEventLogoutSucceeded {
			logoutAudits++
		}
	}
	if logoutAudits != 1 {
		t.Fatalf("logout audit count = %d, want 1", logoutAudits)
	}
}

func TestLogoutMissingOrInvalidSessionReturnsUnauthorized(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	cases := []struct {
		name    string
		cookies []*http.Cookie
		csrf    string
	}{
		{name: "missing cookie", cookies: nil, csrf: "ignored"},
		{name: "unknown token", cookies: []*http.Cookie{{Name: cookie.SessionCookieName, Value: "unknown-token-value"}}, csrf: "ignored"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, logoutPOST(tc.cookies, tc.csrf))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			if len(rec.Result().Cookies()) != 0 {
				t.Fatal("must not clear session cookie without valid auth")
			}
			payload := parseErrorBody(t, rec.Body.Bytes())
			errObj, _ := payload["error"].(map[string]any)
			if errObj["code"] != "UNAUTHORIZED" {
				t.Fatalf("error code = %v", errObj["code"])
			}
		})
	}
}

func TestLogoutMissingCSRFReturnsForbidden(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, _ := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, ""))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not clear session cookie on CSRF failure")
	}

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := router.sessions.GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.RevokedAt != "" {
		t.Fatal("session must not be revoked on CSRF failure")
	}
}

func TestLogoutInvalidCSRFDoesNotRevokeOrAudit(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, _ := loginAndGetCredentials(t, router)

	auditBefore, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, "wrong-csrf-token"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	payload := parseErrorBody(t, rec.Body.Bytes())
	errObj, _ := payload["error"].(map[string]any)
	if errObj["code"] != "FORBIDDEN" || errObj["message"] != "CSRF validation failed" {
		t.Fatalf("error shape = %v", errObj)
	}

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := router.sessions.GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.RevokedAt != "" {
		t.Fatal("session must not be revoked on CSRF failure")
	}

	auditAfter, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(auditAfter) != len(auditBefore) {
		t.Fatal("CSRF failure must not append audit events")
	}
}

func TestLogoutExpiredSessionReturnsUnauthorized(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

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
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range events {
		if event.EventType == appidentity.AuditEventLogoutSucceeded {
			t.Fatal("expired logout must not write logout audit")
		}
	}
}

func TestLogoutDoesNotCreateSession(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	countBefore, err := countSessions(context.Background(), router.store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST(nil, ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
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

type controllableLogoutRevoker struct {
	revoked bool
	err     error
	calls   int
}

func (c *controllableLogoutRevoker) RevokeActiveSessionByTokenHash(
	context.Context,
	string,
	string,
	appidentity.AppendSecurityAuditEventInput,
) (bool, error) {
	c.calls++
	if c.err != nil {
		return false, c.err
	}
	return c.revoked, nil
}

func newIdentityRouterWithCustomLogoutService(t *testing.T, logoutService *appidentity.LogoutService) identityRouter {
	t.Helper()

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
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, sqlite.NewLoginRepository(store.DB()), sqlite.NewSecurityAuditRepository(store.DB()))
	login := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookie.NewPolicy(false),
	})
	logout := handlers.NewLogout(handlers.LogoutDeps{
		Service:      logoutService,
		CookiePolicy: cookie.NewPolicy(false),
	})

	return identityRouter{
		handler:  httpapiNewRouter(store, bootstrap, login, logout),
		store:    store,
		audit:    sqlite.NewSecurityAuditRepository(store.DB()),
		sessions: sqlite.NewSessionRepository(store.DB()),
	}
}

func TestLogoutUseCaseSuccessClearsCookieOnce(t *testing.T) {
	revoker := &controllableLogoutRevoker{revoked: true}
	router := newIdentityRouterWithCustomLogoutService(t, appidentity.NewLogoutService(revoker))
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if revoker.calls != 1 {
		t.Fatalf("logout use case calls = %d, want 1", revoker.calls)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("clear cookie count = %d, want 1", len(cookies))
	}
	if cookies[0].Name != cookie.SessionCookieName || cookies[0].MaxAge != -1 || cookies[0].Value != "" {
		t.Fatalf("cleared cookie = %#v", cookies[0])
	}
}

func TestLogoutUseCaseGenericErrorDoesNotClearCookie(t *testing.T) {
	revoker := &controllableLogoutRevoker{err: errors.New("persist failed")}
	router := newIdentityRouterWithCustomLogoutService(t, appidentity.NewLogoutService(revoker))
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if revoker.calls != 1 {
		t.Fatalf("logout use case calls = %d, want 1", revoker.calls)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("Set-Cookie count = %d, want 0", len(rec.Result().Cookies()))
	}
}

func TestLogoutUseCaseContextCanceledDoesNotClearCookie(t *testing.T) {
	revoker := &controllableLogoutRevoker{err: context.Canceled}
	router := newIdentityRouterWithCustomLogoutService(t, appidentity.NewLogoutService(revoker))
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("status = %d, want 408", rec.Code)
	}
	if revoker.calls != 1 {
		t.Fatalf("logout use case calls = %d, want 1", revoker.calls)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("Set-Cookie count = %d, want 0", len(rec.Result().Cookies()))
	}
}

func TestLogoutUseCaseDeadlineExceededDoesNotClearCookie(t *testing.T) {
	revoker := &controllableLogoutRevoker{err: context.DeadlineExceeded}
	router := newIdentityRouterWithCustomLogoutService(t, appidentity.NewLogoutService(revoker))
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if revoker.calls != 1 {
		t.Fatalf("logout use case calls = %d, want 1", revoker.calls)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("Set-Cookie count = %d, want 0", len(rec.Result().Cookies()))
	}
}

func TestLogoutRevokeFailureReturns500WithoutCookieClear(t *testing.T) {
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
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, sqlite.NewLoginRepository(store.DB()), sqlite.NewSecurityAuditRepository(store.DB()))
	login := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookie.NewPolicy(false),
	})
	logoutService := appidentity.NewLogoutService(&failingLogoutRevoker{})
	logout := handlers.NewLogout(handlers.LogoutDeps{
		Service:      logoutService,
		CookiePolicy: cookie.NewPolicy(false),
	})

	router := identityRouter{
		handler:  httpapiNewRouter(store, bootstrap, login, logout),
		store:    store,
		audit:    sqlite.NewSecurityAuditRepository(store.DB()),
		sessions: sqlite.NewSessionRepository(store.DB()),
	}

	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	auditBefore, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not clear session cookie on revoke failure")
	}
	payload := parseErrorBody(t, rec.Body.Bytes())
	errObj, _ := payload["error"].(map[string]any)
	if errObj["code"] != "INTERNAL_ERROR" {
		t.Fatalf("error code = %v", errObj["code"])
	}

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := sqlite.NewSessionRepository(store.DB()).GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.Session.RevokedAt != "" {
		t.Fatal("session must not be marked revoked when revoke failed")
	}

	auditAfter, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range auditAfter[len(auditBefore):] {
		if event.EventType == appidentity.AuditEventLogoutSucceeded {
			t.Fatal("revoke failure must not write logout success audit")
		}
	}

	resolver := appidentity.NewSessionResolver(sqlite.NewSessionAuthRepository(store.DB()))
	resolved, ok, err := resolver.Resolve(context.Background(), sessionCookie.Value)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if !ok {
		t.Fatal("session must remain active and resolvable after revoke failure")
	}
	if resolved.CSRFTokenHash == "" {
		t.Fatal("resolved session must retain CSRF hash")
	}
}

func httpapiNewRouter(store *sqlite.Store, bootstrap *handlers.Bootstrap, login *handlers.Login, logout *handlers.Logout) http.Handler {
	cfg := config.Config{Env: "development", ReadTimeout: 15 * time.Second}
	resolver := appidentity.NewSessionResolver(sqlite.NewSessionAuthRepository(store.DB()))
	return httpapi.NewRouter(
		cfg,
		slog.Default(),
		health.NewReadiness(store),
		bootstrap,
		login,
		logout,
		nil,
		nil,
		nil,
		nil,
		&httpapi.SessionAuth{Resolver: resolver, Authorizer: ports.NewRBACAuthorizer()},
	)
}

func TestLogoutCSRFHeaderOnly(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	req := logoutPOST([]*http.Cookie{sessionCookie}, "")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.URL.RawQuery = "csrf_token=" + csrfToken

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestLogoutWhitespaceOnlyCSRFRejected(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, _ := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, "   "))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestLogoutOversizedCSRFRejected(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, _ := loginAndGetCredentials(t, router)

	oversized := strings.Repeat("a", appidentity.MaxCSRFHeaderValueLen+1)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, oversized))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestLogoutResponseDoesNotLeakCSRF(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, csrfToken := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), csrfToken) {
		t.Fatal("logout response must not contain CSRF token")
	}
	if strings.Contains(rec.Body.String(), sessionCookie.Value) {
		t.Fatal("logout response must not contain session token")
	}
}

func TestLogoutForbiddenResponseShape(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)
	sessionCookie, _ := loginAndGetCredentials(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, logoutPOST([]*http.Cookie{sessionCookie}, "bad"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	errObj, _ := payload["error"].(map[string]any)
	if errObj["code"] != "FORBIDDEN" {
		t.Fatalf("code = %v", errObj["code"])
	}
}
