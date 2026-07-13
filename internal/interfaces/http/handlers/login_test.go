package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

const testLoginPassword = "valid-password-123"

type identityRouter struct {
	handler  http.Handler
	store    *sqlite.Store
	hasher   *appidentity.PasswordHasher
	users    *sqlite.UserRepository
	sessions *sqlite.SessionRepository
	audit    *sqlite.SecurityAuditRepository
}

func newIdentityRouter(t *testing.T, cookieSecure bool) identityRouter {
	t.Helper()

	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	hasher, err := appidentity.NewPasswordHasher(appidentity.Argon2idConfig{
		Memory:      4096,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
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
	loginRepo := sqlite.NewLoginRepository(store.DB())
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, loginRepo, sqlite.NewSecurityAuditRepository(store.DB()))
	logoutRepo := sqlite.NewLogoutRepository(store.DB())
	logoutService := appidentity.NewLogoutService(logoutRepo)
	cookiePolicy := cookie.NewPolicy(cookieSecure)

	login := handlers.NewLogin(handlers.LoginDeps{Service: loginService, CookiePolicy: cookiePolicy})
	logout := handlers.NewLogout(handlers.LogoutDeps{Service: logoutService, CookiePolicy: cookiePolicy})

	env := "production"
	if !cookieSecure {
		env = "development"
	}

	cfg := config.Config{
		Env:         env,
		Version:     "test",
		BuildCommit: "test",
		ReadTimeout: 15 * time.Second,
	}

	resolver := appidentity.NewSessionResolver(sqlite.NewSessionAuthRepository(store.DB()))
	router := httpapi.NewRouter(
		cfg,
		slog.Default(),
		health.NewReadiness(store),
		bootstrap,
		login,
		logout,
		nil,
		nil,
		&httpapi.SessionAuth{Resolver: resolver, Authorizer: ports.NewRBACAuthorizer()},
	)
	return identityRouter{
		handler:  router,
		store:    store,
		hasher:   hasher,
		users:    userRepo,
		sessions: sqlite.NewSessionRepository(store.DB()),
		audit:    sqlite.NewSecurityAuditRepository(store.DB()),
	}
}

func bootstrapOwner(t *testing.T, router identityRouter) {
	t.Helper()

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d, want 201", rec.Code)
	}
}

func identityPOST(path, body string, cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func loginPOST(body string, cookies []*http.Cookie) *http.Request {
	return identityPOST("/api/v1/identity/login", body, cookies)
}

func logoutPOST(cookies []*http.Cookie, csrfToken string) *http.Request {
	req := identityPOST("/api/v1/identity/logout", "", cookies)
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	return req
}

func loginAndGetCredentials(t *testing.T, router identityRouter) (*http.Cookie, string) {
	t.Helper()

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("login body keys = %v, want csrf_token only", body)
	}
	csrfToken, ok := body["csrf_token"]
	if !ok || csrfToken == "" {
		t.Fatalf("login body = %v", body)
	}

	sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie")
	}
	return sessionCookie, csrfToken
}

func parseErrorBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	return payload
}

func authFailureShape(t *testing.T, body []byte) {
	t.Helper()
	payload := parseErrorBody(t, body)
	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error object: %v", payload)
	}
	if errObj["code"] != "UNAUTHORIZED" || errObj["message"] != "Authentication failed" {
		t.Fatalf("error shape = %v", errObj)
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestLoginValidOwnerSetsCookiesAndSession(t *testing.T) {
	secure := true
	router := newIdentityRouter(t, secure)
	bootstrapOwner(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var loginBody map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if len(loginBody) != 1 || loginBody["csrf_token"] == "" {
		t.Fatalf("login body = %v", loginBody)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}

	sessionCookie := findCookie(cookies, cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie")
	}
	if sessionCookie.Value == "" {
		t.Fatal("empty session cookie value")
	}
	if !sessionCookie.HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	if !sessionCookie.Secure {
		t.Fatal("session cookie must be Secure in production")
	}
	if sessionCookie.Path != "/" {
		t.Fatalf("session path = %q, want /", sessionCookie.Path)
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("session SameSite = %v, want Strict", sessionCookie.SameSite)
	}
	if sessionCookie.Domain != "" {
		t.Fatalf("session Domain must be empty, got %q", sessionCookie.Domain)
	}
	if sessionCookie.MaxAge <= 0 {
		t.Fatalf("session MaxAge = %d, want positive", sessionCookie.MaxAge)
	}

	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := router.sessions.GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}
	if cred.SessionTokenHash != tokenHash {
		t.Fatal("stored session hash mismatch")
	}
	if cred.SessionTokenHash == sessionCookie.Value {
		t.Fatal("raw session token persisted")
	}
	if cred.CSRFTokenHash == "" || cred.CSRFTokenHash == cred.SessionTokenHash {
		t.Fatal("csrf token hash must be persisted separately from session hash")
	}
	owner, err := router.users.GetUserByUsername(context.Background(), "owner")
	if err != nil {
		t.Fatalf("GetUserByUsername() error: %v", err)
	}
	if cred.Session.UserID != owner.User.ID {
		t.Fatalf("session user_id = %q, want %q", cred.Session.UserID, owner.User.ID)
	}
	if cred.Session.ExpiresAt == "" || cred.Session.IdleExpiresAt == "" {
		t.Fatal("session lifecycle timestamps must be persisted")
	}
	if sessionCookie.MaxAge <= 0 {
		t.Fatalf("session MaxAge = %d", sessionCookie.MaxAge)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	foundLoginAudit := false
	for _, event := range events {
		if event.EventType == appidentity.AuditEventLoginSucceeded {
			foundLoginAudit = true
		}
		if strings.Contains(event.MetadataJSON, testLoginPassword) ||
			strings.Contains(event.MetadataJSON, sessionCookie.Value) {
			t.Fatalf("audit leaked secret material: %q", event.MetadataJSON)
		}
	}
	if !foundLoginAudit {
		t.Fatal("missing identity.login.succeeded audit event")
	}

	body := rec.Body.String()
	if strings.Contains(body, testLoginPassword) {
		t.Fatal("response leaked password material")
	}
	if strings.Contains(body, sessionCookie.Value) {
		t.Fatal("response leaked session cookie value")
	}
}

func TestLoginPersistsLifecycleTimestampsAlignedToMaterial(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie")
	}
	tokenHash := appidentity.HashRawToken(sessionCookie.Value)
	cred, err := router.sessions.GetSessionByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() error: %v", err)
	}

	createdAt, err := time.Parse(time.RFC3339, cred.Session.CreatedAt)
	if err != nil {
		t.Fatalf("parse created_at: %v", err)
	}
	if cred.Session.LastSeenAt != cred.Session.CreatedAt {
		t.Fatalf("last_seen_at = %q, want %q", cred.Session.LastSeenAt, cred.Session.CreatedAt)
	}
	if cred.Session.ExpiresAt != appidentity.FormatUTCTime(createdAt.Add(appidentity.DefaultSessionTokenConfig.AbsoluteTTL)) {
		t.Fatalf("expires_at = %q, want %q", cred.Session.ExpiresAt, appidentity.FormatUTCTime(createdAt.Add(appidentity.DefaultSessionTokenConfig.AbsoluteTTL)))
	}
	if cred.Session.IdleExpiresAt != appidentity.FormatUTCTime(createdAt.Add(appidentity.DefaultSessionTokenConfig.IdleTTL)) {
		t.Fatalf("idle_expires_at = %q, want %q", cred.Session.IdleExpiresAt, appidentity.FormatUTCTime(createdAt.Add(appidentity.DefaultSessionTokenConfig.IdleTTL)))
	}
}

func TestLoginDevelopmentCookieSecureDefault(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie")
	}
	if sessionCookie.Secure {
		t.Fatal("development default must allow insecure cookies")
	}
}

func TestLoginCredentialFailuresIndistinguishable(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	hash, err := router.hasher.HashPassword(context.Background(), testLoginPassword)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if err := router.users.CreateUser(context.Background(), appidentity.CreateUserInput{
		ID:           domainidentity.UserID("user-disabled"),
		Username:     "disabled-user",
		PasswordHash: hash,
		Role:         domainidentity.RoleUser,
		Status:       appidentity.UserStatusDisabled,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	if err := router.users.CreateUser(context.Background(), appidentity.CreateUserInput{
		ID:           domainidentity.UserID("user-bad-hash"),
		Username:     "bad-hash-user",
		PasswordHash: "not-a-valid-hash",
		Role:         domainidentity.RoleUser,
		Status:       appidentity.UserStatusActive,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	cases := []struct {
		name string
		body string
	}{
		{name: "unknown username", body: `{"username":"missing","password":"` + testLoginPassword + `"}`},
		{name: "wrong password", body: `{"username":"owner","password":"wrong-password-999"}`},
		{name: "disabled user", body: `{"username":"disabled-user","password":"` + testLoginPassword + `"}`},
		{name: "malformed stored hash", body: `{"username":"bad-hash-user","password":"` + testLoginPassword + `"}`},
	}

	var referenceStatus int
	var referenceBody []byte

	for i, tc := range cases {
		rec := httptest.NewRecorder()
		router.handler.ServeHTTP(rec, loginPOST(tc.body, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s status = %d, want 401", tc.name, rec.Code)
		}
		if len(rec.Result().Cookies()) != 0 {
			t.Fatalf("%s set cookies on failure", tc.name)
		}
		authFailureShape(t, rec.Body.Bytes())
		if i == 0 {
			referenceStatus = rec.Code
			referenceBody = append([]byte(nil), rec.Body.Bytes()...)
			referenceBody = normalizeAuthFailureBody(referenceBody)
		} else {
			if rec.Code != referenceStatus {
				t.Fatalf("%s status differs", tc.name)
			}
			if string(normalizeAuthFailureBody(rec.Body.Bytes())) != string(referenceBody) {
				t.Fatalf("%s body differs: %s", tc.name, rec.Body.String())
			}
		}
	}

	count, err := countSessions(context.Background(), router.store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}
	if count != 0 {
		t.Fatalf("sessions after failures = %d, want 0", count)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	failureAudits := 0
	for _, event := range events {
		if event.EventType == appidentity.AuditEventLoginFailure {
			failureAudits++
		}
		if strings.Contains(event.MetadataJSON, testLoginPassword) {
			t.Fatalf("audit leaked password material: %q", event.MetadataJSON)
		}
	}
	if failureAudits != len(cases) {
		t.Fatalf("failure audit count = %d, want %d", failureAudits, len(cases))
	}
}

func normalizeAuthFailureBody(body []byte) []byte {
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	if errObj, ok := payload["error"].(map[string]any); ok {
		delete(errObj, "request_id")
	}
	out, _ := json.Marshal(payload)
	return out
}

func TestLoginRejectsInvalidRequests(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	cases := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "unknown field",
			req:  loginPOST(`{"username":"owner","password":"`+testLoginPassword+`","role":"owner"}`, nil),
		},
		{
			name: "invalid json",
			req:  loginPOST(`{"username":"owner"`, nil),
		},
		{
			name: "missing content type",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/login", strings.NewReader(`{"username":"owner","password":"x"}`))
				req.RemoteAddr = "127.0.0.1:8080"
				return req
			}(),
		},
		{
			name: "empty username",
			req:  loginPOST(`{"username":"","password":"`+testLoginPassword+`"}`, nil),
		},
		{
			name: "empty password",
			req:  loginPOST(`{"username":"owner","password":""}`, nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, tc.req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if len(rec.Result().Cookies()) != 0 {
				t.Fatal("must not set cookies")
			}
		})
	}

	largePassword := strings.Repeat("a", 1025)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+largePassword+`"}`, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized password status = %d, want 400", rec.Code)
	}

	rec = httptest.NewRecorder()
	oversizedBody := `{"username":"owner","password":"` + strings.Repeat("a", 9*1024) + `"}`
	router.handler.ServeHTTP(rec, loginPOST(oversizedBody, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized body status = %d, want 400", rec.Code)
	}

	count, err := countSessions(context.Background(), router.store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}
	if count != 0 {
		t.Fatalf("malformed login must not create sessions: count = %d", count)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range events {
		if event.EventType == appidentity.AuditEventLoginFailure {
			t.Fatalf("malformed login must not write failure audit: %s", event.EventType)
		}
	}
}

type failingLoginAuditStore struct{}

func (f failingLoginAuditStore) AppendSecurityAuditEvent(context.Context, appidentity.AppendSecurityAuditEventInput) error {
	return errors.New("audit persist failed")
}

func (f failingLoginAuditStore) ListSecurityAuditEvents(context.Context, appidentity.ListSecurityAuditEventsInput) ([]appidentity.SecurityAuditEvent, error) {
	return nil, nil
}

func TestLoginFailureAuditPersistenceErrorReturns500(t *testing.T) {
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
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, sqlite.NewLoginRepository(store.DB()), failingLoginAuditStore{})
	login := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookie.NewPolicy(false),
	})

	cfg := config.Config{Env: "development", ReadTimeout: 15 * time.Second}
	router := httpapi.NewRouter(cfg, slog.Default(), health.NewReadiness(store), bootstrap, login, nil, nil, nil, nil)

	recBootstrap := httptest.NewRecorder()
	router.ServeHTTP(recBootstrap, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if recBootstrap.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d", recBootstrap.Code)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"wrong-password"}`, nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not set cookies when failure audit persist fails")
	}
}

func TestLoginNeedsRehashUpdatesHashAndCreatesSession(t *testing.T) {
	router := newIdentityRouter(t, false)

	oldHasher, err := appidentity.NewPasswordHasher(appidentity.Argon2idConfig{
		Memory:      2048,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher(old) error: %v", err)
	}
	oldHash, err := oldHasher.HashPassword(context.Background(), testLoginPassword)
	if err != nil {
		t.Fatalf("HashPassword(old) error: %v", err)
	}
	if err := router.users.CreateUser(context.Background(), appidentity.CreateUserInput{
		ID:           domainidentity.UserID("user-rehash"),
		Username:     "rehash-user",
		PasswordHash: oldHash,
		Role:         domainidentity.RoleOwner,
		Status:       appidentity.UserStatusActive,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"rehash-user","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	cred, err := router.users.GetUserByUsername(context.Background(), "rehash-user")
	if err != nil {
		t.Fatalf("GetUserByUsername() error: %v", err)
	}
	if cred.PasswordHash == oldHash {
		t.Fatal("password hash was not updated")
	}
	valid, needsRehash, err := router.hasher.VerifyPassword(context.Background(), testLoginPassword, cred.PasswordHash)
	if err != nil || !valid || needsRehash {
		t.Fatalf("updated hash verify = (%v,%v,%v)", valid, needsRehash, err)
	}

	count, err := countSessions(context.Background(), router.store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}
	if count != 1 {
		t.Fatalf("session count = %d, want 1", count)
	}
}

type failingPasswordUpdater struct {
	*sqlite.UserRepository
}

func (f *failingPasswordUpdater) UpdateUserPasswordHash(ctx context.Context, id domainidentity.UserID, passwordHash string) error {
	return errors.New("update failed")
}

func TestLoginRehashFailureSetsNoSession(t *testing.T) {
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
	oldHasher, err := appidentity.NewPasswordHasher(appidentity.Argon2idConfig{
		Memory: 2048, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher(old) error: %v", err)
	}

	userRepo := sqlite.NewUserRepository(store.DB())
	oldHash, err := oldHasher.HashPassword(context.Background(), testLoginPassword)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if err := userRepo.CreateUser(context.Background(), appidentity.CreateUserInput{
		ID: domainidentity.UserID("user-rehash-fail"), Username: "rehash-fail",
		PasswordHash: oldHash, Role: domainidentity.RoleOwner, Status: appidentity.UserStatusActive,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	sessionTokens, err := appidentity.NewSessionTokenService(appidentity.DefaultSessionTokenConfig)
	if err != nil {
		t.Fatalf("NewSessionTokenService() error: %v", err)
	}
	failingUsers := &failingPasswordUpdater{UserRepository: userRepo}
	loginService := appidentity.NewLoginService(failingUsers, hasher, sessionTokens, sqlite.NewLoginRepository(store.DB()), sqlite.NewSecurityAuditRepository(store.DB()))
	login := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookie.NewPolicy(false),
	})

	cfg := config.Config{Env: "development", ReadTimeout: 15 * time.Second}
	router := httpapi.NewRouter(cfg, slog.Default(), health.NewReadiness(store), nil, login, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, loginPOST(`{"username":"rehash-fail","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not set cookies")
	}
	count, err := countSessions(context.Background(), store.DB())
	if err != nil {
		t.Fatalf("countSessions() error: %v", err)
	}
	if count != 0 {
		t.Fatalf("session count = %d, want 0", count)
	}
}

type failingLoginCreator struct{}

func (f *failingLoginCreator) CreateSessionWithAudit(
	ctx context.Context,
	session appidentity.CreateSessionInput,
	audit appidentity.AppendSecurityAuditEventInput,
) error {
	return errors.New("persist failed")
}

func TestLoginSessionPersistenceFailureSetsNoCookie(t *testing.T) {
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
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, &failingLoginCreator{}, sqlite.NewSecurityAuditRepository(store.DB()))
	login := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookie.NewPolicy(false),
	})

	cfg := config.Config{Env: "development", ReadTimeout: 15 * time.Second}
	router := httpapi.NewRouter(cfg, slog.Default(), health.NewReadiness(store), bootstrap, login, nil, nil, nil, nil)

	recBootstrap := httptest.NewRecorder()
	router.ServeHTTP(recBootstrap, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if recBootstrap.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d", recBootstrap.Code)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not set cookies")
	}
}

func TestLoginRequestCancellationFailsSafely(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("status = %d, want 408", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not set cookies")
	}
}

func TestLoginConcurrentCreatesSeparateSessions(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	const workers = 4
	var wg sync.WaitGroup
	hashes := make(chan string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
				return
			}
			sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
			if sessionCookie == nil {
				t.Error("missing session cookie")
				return
			}
			hashes <- appidentity.HashRawToken(sessionCookie.Value)
		}()
	}
	wg.Wait()
	close(hashes)

	seen := make(map[string]struct{})
	for hash := range hashes {
		if _, ok := seen[hash]; ok {
			t.Fatalf("duplicate session hash %q", hash)
		}
		seen[hash] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("unique sessions = %d, want %d", len(seen), workers)
	}
}

func countSessions(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&count)
	return count, err
}

func TestLoginStoredPasswordIsArgon2id(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	cred, err := router.users.GetUserByUsername(context.Background(), "owner")
	if err != nil {
		t.Fatalf("GetUserByUsername() error: %v", err)
	}
	if !strings.HasPrefix(cred.PasswordHash, "$argon2id$") {
		t.Fatalf("password hash prefix = %q", cred.PasswordHash[:min(20, len(cred.PasswordHash))])
	}
	if strings.Contains(rec.Body.String(), cred.PasswordHash) {
		t.Fatal("response leaked password hash")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestLoginWrongContentType(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/login", strings.NewReader(`{"username":"owner","password":"x"}`))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestLoginSecondJSONValueRejected(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	body := `{"username":"owner","password":"` + testLoginPassword + `"}{}`
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(body, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestLoginDoesNotReadResponseSecrets(t *testing.T) {
	router := newIdentityRouter(t, false)
	bootstrapOwner(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if len(payload) != 1 || payload["csrf_token"] == "" {
		t.Fatalf("body = %#v, want only csrf_token", payload)
	}
	sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie")
	}
	if payload["csrf_token"] == sessionCookie.Value {
		t.Fatal("csrf_token must not equal session cookie value")
	}
	if strings.Contains(rec.Body.String(), sessionCookie.Value) {
		t.Fatal("response must not contain raw session token")
	}
}
