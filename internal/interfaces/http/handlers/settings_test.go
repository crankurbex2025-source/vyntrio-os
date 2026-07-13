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
	"sync/atomic"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

const (
	settingsPath            = "/api/v1/settings"
	settingsFutureExpiry    = "2099-01-01T00:00:00Z"
	settingsTestVersion     = "0.2.0-test"
	settingsTestEnvironment = "development"
)

type settingsRouter struct {
	handler  http.Handler
	store    *sqlite.Store
	users    *sqlite.UserRepository
	sessions *sqlite.SessionRepository
	audit    *sqlite.SecurityAuditRepository
	view     appsettings.PublicView
}

type settingsRouterOpts struct {
	authorizer ports.Authorizer
	resolver   *appidentity.SessionResolver
	store      *sqlite.Store
	routerOpts []httpapi.RouterOption
}

func newSettingsRouter(t *testing.T, opts settingsRouterOpts) settingsRouter {
	t.Helper()
	store := opts.store
	if store == nil {
		dir := t.TempDir()
		var err error
		store, err = sqlite.Open(context.Background(), dir)
		if err != nil {
			t.Fatalf("Open() error: %v", err)
		}
		t.Cleanup(func() { _ = store.Close() })
	}
	return buildSettingsRouter(t, store, opts)
}

func buildSettingsRouter(t *testing.T, store *sqlite.Store, opts settingsRouterOpts) settingsRouter {
	t.Helper()

	ctx := context.Background()
	settingsRepo := sqlite.NewSettingsRepository(store.DB())
	sysSettings, err := appsettings.NewReader(settingsRepo).LoadSystemSettings(ctx)
	if err != nil {
		t.Fatalf("LoadSystemSettings() error: %v", err)
	}
	snapshot := appsettings.NewSnapshot(sysSettings)
	view := appsettings.NewPublicView(snapshot, settingsTestVersion, settingsTestEnvironment)
	settingsLoader := appsettings.NewPublicSettingsLoader(settingsRepo, settingsTestVersion, settingsTestEnvironment)

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
	cookiePolicy := cookie.NewPolicy(false)

	login := handlers.NewLogin(handlers.LoginDeps{Service: loginService, CookiePolicy: cookiePolicy})
	logout := handlers.NewLogout(handlers.LogoutDeps{Service: logoutService, CookiePolicy: cookiePolicy})
	settings := handlers.NewSettings(handlers.SettingsDeps{Loader: settingsLoader})
	instanceDisplayNameRepo := sqlite.NewInstanceDisplayNameRepository(store.DB())
	updateInstanceService := appsettings.NewUpdateInstanceDisplayNameService(instanceDisplayNameRepo)
	updateInstance := handlers.NewUpdateInstanceSettings(handlers.UpdateInstanceSettingsDeps{Service: updateInstanceService})

	resolver := opts.resolver
	if resolver == nil {
		resolver = appidentity.NewSessionResolver(sqlite.NewSessionAuthRepository(store.DB()))
	}
	authorizer := opts.authorizer
	if authorizer == nil {
		authorizer = ports.NewRBACAuthorizer()
	}

	cfg := config.Config{
		Env:         settingsTestEnvironment,
		Version:     settingsTestVersion,
		BuildCommit: "must-not-leak",
		ReadTimeout: 15 * time.Second,
	}

	router := httpapi.NewRouter(
		cfg,
		slog.Default(),
		health.NewReadiness(store),
		bootstrap,
		login,
		logout,
		settings,
		updateInstance,
		&httpapi.SessionAuth{
			Resolver:   resolver,
			Authorizer: authorizer,
		},
		opts.routerOpts...,
	)

	return settingsRouter{
		handler:  router,
		store:    store,
		users:    userRepo,
		sessions: sqlite.NewSessionRepository(store.DB()),
		audit:    sqlite.NewSecurityAuditRepository(store.DB()),
		view:     view,
	}
}

func settingsGET(cookies []*http.Cookie, query string) *http.Request {
	path := settingsPath
	if query != "" {
		path += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "127.0.0.1:8080"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func ownerSessionCookie(t *testing.T, router settingsRouter) *http.Cookie {
	t.Helper()

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d, want 201", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.handler.ServeHTTP(rec, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200", rec.Code)
	}
	sessionCookie := findCookie(rec.Result().Cookies(), cookie.SessionCookieName)
	if sessionCookie == nil {
		t.Fatal("missing session cookie after login")
	}
	return sessionCookie
}

func parseSettingsErrorCode(t *testing.T, body []byte) string {
	t.Helper()
	payload := parseErrorBody(t, body)
	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error object: %v", payload)
	}
	code, _ := errObj["code"].(string)
	return code
}

func assertNoSensitiveSettingsFields(t *testing.T, body []byte) {
	t.Helper()
	lower := strings.ToLower(string(body))
	for _, forbidden := range []string{
		"password", "token", "hash", "csrf", "session", "userid", "user_id",
		"role", "principal", "commit", "datadir", "data_dir", "path", "port",
		"host", "bind", "timezone", "127.0.0.1", "sqlite", "audit",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("response leaked forbidden substring %q: %s", forbidden, body)
		}
	}
}

func countSettingsRows(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM settings`).Scan(&count); err != nil {
		t.Fatalf("count settings rows: %v", err)
	}
	return count
}

func createSettingsSession(t *testing.T, router settingsRouter, rawToken string, userStatus appidentity.UserStatus) *http.Cookie {
	t.Helper()
	ctx := context.Background()

	userID := domainidentity.UserID("settings-user-1")
	if err := router.users.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           userID,
		Username:     "settings-user",
		PasswordHash: "hash",
		Role:         domainidentity.RoleOwner,
		Status:       userStatus,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	if err := router.sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               "settings-sess-1",
		UserID:           userID,
		SessionTokenHash: appidentity.HashRawToken(rawToken),
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        "2026-07-13T10:00:00Z",
		LastSeenAt:       "2026-07-13T10:00:00Z",
		ExpiresAt:        settingsFutureExpiry,
		IdleExpiresAt:    settingsFutureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}
	return &http.Cookie{Name: cookie.SessionCookieName, Value: rawToken}
}

func createRoleSettingsSession(
	t *testing.T,
	router settingsRouter,
	userID domainidentity.UserID,
	username string,
	role domainidentity.Role,
	sessionID string,
	rawToken string,
) *http.Cookie {
	t.Helper()
	ctx := context.Background()

	if err := router.users.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           userID,
		Username:     username,
		PasswordHash: "hash",
		Role:         role,
		Status:       appidentity.UserStatusActive,
	}); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	if err := router.sessions.CreateSession(ctx, appidentity.CreateSessionInput{
		ID:               sessionID,
		UserID:           userID,
		SessionTokenHash: appidentity.HashRawToken(rawToken),
		CSRFTokenHash:    "csrf-hash",
		CreatedAt:        "2026-07-13T10:00:00Z",
		LastSeenAt:       "2026-07-13T10:00:00Z",
		ExpiresAt:        settingsFutureExpiry,
		IdleExpiresAt:    settingsFutureExpiry,
	}); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}
	return &http.Cookie{Name: cookie.SessionCookieName, Value: rawToken}
}

type denyAuthorizer struct{}

func (denyAuthorizer) Authorize(domainidentity.Principal, domainidentity.Permission) error {
	return domainidentity.ErrForbidden
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

type failingSessionAuthStore struct{}

func (failingSessionAuthStore) GetSessionAuthByTokenHash(
	context.Context,
	string,
) (appidentity.SessionAuthRecord, error) {
	return appidentity.SessionAuthRecord{}, errors.New("database unavailable")
}

func TestSettingsValidOwnerSessionReturnsSafeDTO(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)

	settingsBefore := countSettingsRows(t, router.store.DB())
	auditBefore, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 100})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("Set-Cookie count = %d, want 0", len(rec.Result().Cookies()))
	}

	want := router.view.Response()
	var got appsettings.PublicSettingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if got != want {
		t.Fatalf("response = %+v, want %+v", got, want)
	}
	assertNoSensitiveSettingsFields(t, rec.Body.Bytes())

	settingsAfter := countSettingsRows(t, router.store.DB())
	if settingsAfter != settingsBefore {
		t.Fatalf("settings row count changed: before=%d after=%d", settingsBefore, settingsAfter)
	}
	auditAfter, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 100})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(auditAfter) != len(auditBefore) {
		t.Fatalf("audit event count changed: before=%d after=%d", len(auditBefore), len(auditAfter))
	}
}

func TestSettingsMissingSessionCookieReturns401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET(nil, ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"instance"`) {
		t.Fatal("settings handler response leaked on 401")
	}
}

func TestSettingsInvalidSessionCookiesReturn401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	ownerSessionCookie(t, router)

	cases := []struct {
		name  string
		value string
	}{
		{name: "empty value", value: ""},
		{name: "unknown token", value: "unknown-session-token-value"},
		{name: "oversized value", value: strings.Repeat("a", appidentity.MaxSessionCookieValueLen+1)},
		{name: "malformed value", value: "not-a-valid-base64!!!"},
	}

	var referenceBody string
	for i, tc := range cases {
		rec := httptest.NewRecorder()
		router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{{Name: cookie.SessionCookieName, Value: tc.value}}, ""))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s status = %d", tc.name, rec.Code)
		}
		if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
			t.Fatalf("%s body = %s", tc.name, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), `"instance"`) {
			t.Fatalf("%s settings handler response leaked", tc.name)
		}
		normalized := normalizeSettingsErrorBody(rec.Body.Bytes())
		if i == 0 {
			referenceBody = normalized
		} else if normalized != referenceBody {
			t.Fatalf("%s body differs from reference: %s", tc.name, rec.Body.String())
		}
	}
}

func normalizeSettingsErrorBody(body []byte) string {
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	if errObj, ok := payload["error"].(map[string]any); ok {
		delete(errObj, "request_id")
	}
	out, _ := json.Marshal(payload)
	return string(out)
}

func TestSettingsRevokedExpiredDisabledSessionsReturn401(t *testing.T) {
	cases := []struct {
		name  string
		token string
		setup func(t *testing.T, router settingsRouter)
	}{
		{
			name:  "revoked session",
			token: "revoked-settings-token",
			setup: func(t *testing.T, router settingsRouter) {
				t.Helper()
				createSettingsSession(t, router, "revoked-settings-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET revoked_at = ? WHERE id = ?`,
					"2026-07-13T12:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("revoke session: %v", err)
				}
			},
		},
		{
			name:  "absolute expired session",
			token: "expired-settings-token",
			setup: func(t *testing.T, router settingsRouter) {
				t.Helper()
				createSettingsSession(t, router, "expired-settings-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET expires_at = ?, idle_expires_at = ? WHERE id = ?`,
					"2000-01-01T00:00:00Z",
					"2000-01-01T00:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("expire session: %v", err)
				}
			},
		},
		{
			name:  "idle expired session",
			token: "idle-expired-settings-token",
			setup: func(t *testing.T, router settingsRouter) {
				t.Helper()
				createSettingsSession(t, router, "idle-expired-settings-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET idle_expires_at = ? WHERE id = ?`,
					"2000-01-01T00:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("idle expire session: %v", err)
				}
			},
		},
		{
			name:  "disabled user",
			token: "disabled-user-settings-token",
			setup: func(t *testing.T, router settingsRouter) {
				t.Helper()
				createSettingsSession(t, router, "disabled-user-settings-token", appidentity.UserStatusDisabled)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			local := newSettingsRouter(t, settingsRouterOpts{})
			tc.setup(t, local)
			sessionCookie := &http.Cookie{Name: cookie.SessionCookieName, Value: tc.token}

			rec := httptest.NewRecorder()
			local.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{sessionCookie}, ""))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
				t.Fatalf("body = %s", rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), `"instance"`) {
				t.Fatal("settings handler response leaked")
			}
		})
	}
}

func TestSettingsAuthenticatedWithoutPermissionReturns403(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{authorizer: denyAuthorizer{}})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"instance"`) {
		t.Fatal("settings handler response leaked on 403")
	}
}

func TestSettingsResolverFailureReturns500(t *testing.T) {
	resolver := appidentity.NewSessionResolver(failingSessionAuthStore{})
	router := newSettingsRouter(t, settingsRouterOpts{resolver: resolver})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{{Name: cookie.SessionCookieName, Value: "token"}}, ""))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "INTERNAL_ERROR" {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"instance"`) {
		t.Fatal("settings handler response leaked on 500")
	}
}

func TestSettingsContextCancellationReturns408(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := createSettingsSession(t, router, "cancel-settings-token", appidentity.UserStatusActive)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := settingsGET([]*http.Cookie{sessionCookie}, "").WithContext(ctx)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("status = %d", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "REQUEST_TIMEOUT" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestSettingsMiddlewareResolvesSessionOnce(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	counter := &countingSessionAuthStore{inner: sqlite.NewSessionAuthRepository(store.DB())}
	router := newSettingsRouter(t, settingsRouterOpts{
		store:    store,
		resolver: appidentity.NewSessionResolver(counter),
	})
	sessionCookie := createSettingsSession(t, router, "resolve-once-token", appidentity.UserStatusActive)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if counter.calls.Load() != 1 {
		t.Fatalf("lookup calls = %d, want 1", counter.calls.Load())
	}
}

func TestSettingsWrongHTTPMethodReturns405(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)

	req := httptest.NewRequest(http.MethodPost, settingsPath, strings.NewReader(`{"ignored":true}`))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "METHOD_NOT_ALLOWED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestSettingsRequestInputsDoNotInfluenceAuthorizationOrResponse(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, settingsPath+"?debug=1&role=owner", strings.NewReader(`{"role":"owner"}`))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Debug", "1")
	req.AddCookie(sessionCookie)
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	want := router.view.Response()
	var got appsettings.PublicSettingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if got != want {
		t.Fatalf("response = %+v, want %+v", got, want)
	}

	rec = httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{{Name: cookie.SessionCookieName, Value: "unknown-token"}}, "role=owner"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", rec.Code)
	}
}

func TestSettingsAuthMiddlewareOnlyAttachedToSettingsRoute(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	counter := &countingSessionAuthStore{inner: sqlite.NewSessionAuthRepository(store.DB())}
	router := newSettingsRouter(t, settingsRouterOpts{
		store:    store,
		resolver: appidentity.NewSessionResolver(counter),
	})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("version status = %d", rec.Code)
	}
	if counter.calls.Load() != 0 {
		t.Fatalf("version lookup calls = %d, want 0", counter.calls.Load())
	}

	rec = httptest.NewRecorder()
	router.handler.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d", rec.Code)
	}
	if counter.calls.Load() != 0 {
		t.Fatalf("bootstrap lookup calls = %d, want 0", counter.calls.Load())
	}
}

func TestSettingsRealRBACAuthorizationMatrix(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	cases := []struct {
		name       string
		userID     domainidentity.UserID
		username   string
		role       domainidentity.Role
		sessionID  string
		rawToken   string
		wantStatus int
	}{
		{
			name:       "owner",
			userID:     domainidentity.UserID("settings-owner"),
			username:   "owner-rbac",
			role:       domainidentity.RoleOwner,
			sessionID:  "settings-owner-sess",
			rawToken:   "owner-settings-rbac-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "administrator",
			userID:     domainidentity.UserID("settings-administrator"),
			username:   "administrator-rbac",
			role:       domainidentity.RoleAdministrator,
			sessionID:  "settings-administrator-sess",
			rawToken:   "administrator-settings-rbac-token",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "operator",
			userID:     domainidentity.UserID("settings-operator"),
			username:   "operator-rbac",
			role:       domainidentity.RoleOperator,
			sessionID:  "settings-operator-sess",
			rawToken:   "operator-settings-rbac-token",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "user",
			userID:     domainidentity.UserID("settings-user"),
			username:   "user-rbac",
			role:       domainidentity.RoleUser,
			sessionID:  "settings-user-sess",
			rawToken:   "user-settings-rbac-token",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "read_only",
			userID:     domainidentity.UserID("settings-read-only"),
			username:   "read-only-rbac",
			role:       domainidentity.RoleReadOnly,
			sessionID:  "settings-read-only-sess",
			rawToken:   "read-only-settings-rbac-token",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionCookie := createRoleSettingsSession(
				t,
				router,
				tc.userID,
				tc.username,
				tc.role,
				tc.sessionID,
				tc.rawToken,
			)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, settingsGET([]*http.Cookie{sessionCookie}, ""))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK {
				want := router.view.Response()
				var got appsettings.PublicSettingsResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("json.Unmarshal() error: %v", err)
				}
				if got != want {
					t.Fatalf("response = %+v, want %+v", got, want)
				}
				return
			}
			if tc.wantStatus == http.StatusForbidden {
				if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
					t.Fatalf("body = %s", rec.Body.String())
				}
			}
			if strings.Contains(rec.Body.String(), `"instance"`) {
				t.Fatal("settings handler response leaked")
			}
		})
	}
}

func TestSettingsHandlerIgnoresRequestBody(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)

	req := httptest.NewRequest(http.MethodGet, settingsPath, strings.NewReader(`{"ignored":true}`))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
