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
	"testing"
	"time"
	"unicode/utf8"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

const instanceSettingsPath = "/api/v1/settings/instance"

func instancePATCH(body string, cookies []*http.Cookie, csrfToken string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, instanceSettingsPath, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/json")
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func instancePATCHAtPath(path, body string, cookies []*http.Cookie, csrfToken string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/json")
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func getHostname(t *testing.T, db *sql.DB) string {
	t.Helper()
	var value string
	if err := db.QueryRowContext(context.Background(),
		`SELECT value FROM settings WHERE namespace = 'system' AND key = 'hostname'`).Scan(&value); err != nil {
		t.Fatalf("query hostname: %v", err)
	}
	return value
}

func getHostnameUpdatedAt(t *testing.T, db *sql.DB) string {
	t.Helper()
	var updatedAt string
	if err := db.QueryRowContext(context.Background(),
		`SELECT updated_at FROM settings WHERE namespace = 'system' AND key = 'hostname'`).Scan(&updatedAt); err != nil {
		t.Fatalf("query updated_at: %v", err)
	}
	return updatedAt
}

func countInstanceDisplayNameAudits(t *testing.T, audit *sqlite.SecurityAuditRepository) int {
	t.Helper()
	events, err := audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 100})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	count := 0
	for _, event := range events {
		if event.EventType == appsettings.AuditEventInstanceDisplayNameUpdated {
			count++
		}
	}
	return count
}

func parseSettingsErrorMessage(t *testing.T, body []byte) string {
	t.Helper()
	payload := parseErrorBody(t, body)
	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error object: %v", payload)
	}
	message, _ := errObj["message"].(string)
	return message
}

func assertNoSetCookie(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("Set-Cookie count = %d, want 0", len(rec.Result().Cookies()))
	}
}

func assertResponseDoesNotContainToken(t *testing.T, rec *httptest.ResponseRecorder, token string) {
	t.Helper()
	if token == "" {
		return
	}
	if strings.Contains(rec.Body.String(), token) {
		t.Fatalf("response body leaked token value %q", token)
	}
	for name, values := range rec.Result().Header {
		for _, v := range values {
			if strings.Contains(v, token) {
				t.Fatalf("response header %q leaked token value", name)
			}
		}
	}
}

func TestUpdateInstanceDisplayNameOwnerSuccess(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler, store: router.store, users: router.users, sessions: router.sessions, audit: router.audit})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"Vyntrio Home"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not set cookies")
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if len(body) != 1 || body["display_name"] != "Vyntrio Home" {
		t.Fatalf("body = %v", body)
	}
	if getHostname(t, router.store.DB()) != "Vyntrio Home" {
		t.Fatal("hostname not persisted")
	}
	if countInstanceDisplayNameAudits(t, router.audit) != 1 {
		t.Fatal("expected one success audit")
	}
}

func TestUpdateInstanceDisplayNameNonOwnerForbidden(t *testing.T) {
	roles := []struct {
		name string
		role domainidentity.Role
	}{
		{name: "administrator", role: domainidentity.RoleAdministrator},
		{name: "operator", role: domainidentity.RoleOperator},
		{name: "user", role: domainidentity.RoleUser},
		{name: "readonly", role: domainidentity.RoleReadOnly},
	}

	for _, tc := range roles {
		t.Run(tc.name, func(t *testing.T) {
			router := newSettingsRouter(t, settingsRouterOpts{})
			rawToken := "token-" + tc.name
			sessionCookie := createRoleSettingsSession(t, router, domainidentity.UserID("user-"+tc.name), tc.name, tc.role, "sess-"+tc.name, rawToken)
			hostnameBefore := getHostname(t, router.store.DB())
			auditBefore := countInstanceDisplayNameAudits(t, router.audit)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"Denied Name"}`, []*http.Cookie{sessionCookie}, "any-csrf"))
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", rec.Code)
			}
			if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
				t.Fatalf("body = %s", rec.Body.String())
			}
			if getHostname(t, router.store.DB()) != hostnameBefore {
				t.Fatal("hostname must not change")
			}
			if countInstanceDisplayNameAudits(t, router.audit) != auditBefore {
				t.Fatal("must not append audit")
			}
		})
	}
}

func TestUpdateInstanceDisplayNameMissingSessionUnauthorized(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	hostnameBefore := getHostname(t, router.store.DB())

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"X"}`, nil, "csrf"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("must not set cookies")
	}
	if getHostname(t, router.store.DB()) != hostnameBefore {
		t.Fatal("hostname must not change")
	}
}

func TestUpdateInstanceDisplayNameInvalidCSRFForbidden(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, _ := loginAndGetCredentials(t, identityRouter{handler: router.handler})
	hostnameBefore := getHostname(t, router.store.DB())

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"X"}`, []*http.Cookie{sessionCookie}, "wrong-csrf"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
	if getHostname(t, router.store.DB()) != hostnameBefore {
		t.Fatal("hostname must not change")
	}
}

func TestUpdateInstanceDisplayNameValidationFailures(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "malformed json", body: `{"display_name":`},
		{name: "multiple values", body: `{"display_name":"A"}{}`},
		{name: "missing field", body: `{}`},
		{name: "unknown field", body: `{"display_name":"A","extra":"x"}`},
		{name: "empty name", body: `{"display_name":""}`},
		{name: "whitespace name", body: `{"display_name":"   "}`},
		{name: "newline", body: "{\"display_name\":\"bad\\nname\"}"},
		{name: "tab", body: "{\"display_name\":\"bad\\tname\"}"},
		{name: "oversized", body: `{"display_name":"` + strings.Repeat("a", setting.MaxInstanceDisplayNameRunes+1) + `"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := newSettingsRouter(t, settingsRouterOpts{})
			bootstrapOwner(t, identityRouter{handler: router.handler})
			sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})
			hostnameBefore := getHostname(t, router.store.DB())
			auditBefore := countInstanceDisplayNameAudits(t, router.audit)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, instancePATCH(tc.body, []*http.Cookie{sessionCookie}, csrfToken))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if getHostname(t, router.store.DB()) != hostnameBefore {
				t.Fatal("hostname must not change")
			}
			if countInstanceDisplayNameAudits(t, router.audit) != auditBefore {
				t.Fatal("must not append audit")
			}
		})
	}
}

func TestUpdateInstanceDisplayNameTrimsWhitespace(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"  Trimmed Name  "}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if getHostname(t, router.store.DB()) != "Trimmed Name" {
		t.Fatalf("hostname = %q", getHostname(t, router.store.DB()))
	}
}

func TestUpdateInstanceDisplayNameNoOpSkipsWriteAndAudit(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec1 := httptest.NewRecorder()
	router.handler.ServeHTTP(rec1, instancePATCH(`{"display_name":"No Op Name"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec1.Code != http.StatusOK {
		t.Fatalf("first status = %d", rec1.Code)
	}
	updatedAt := getHostnameUpdatedAt(t, router.store.DB())
	auditCount := countInstanceDisplayNameAudits(t, router.audit)

	rec2 := httptest.NewRecorder()
	router.handler.ServeHTTP(rec2, instancePATCH(`{"display_name":"No Op Name"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec2.Code != http.StatusOK {
		t.Fatalf("second status = %d", rec2.Code)
	}
	if getHostnameUpdatedAt(t, router.store.DB()) != updatedAt {
		t.Fatal("updated_at must not change on no-op")
	}
	if countInstanceDisplayNameAudits(t, router.audit) != auditCount {
		t.Fatal("no-op must not append audit")
	}
}

func TestUpdateInstanceDisplayNameGETReflectsChange(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	recPatch := httptest.NewRecorder()
	router.handler.ServeHTTP(recPatch, instancePATCH(`{"display_name":"Updated Appliance"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if recPatch.Code != http.StatusOK {
		t.Fatalf("patch status = %d", recPatch.Code)
	}

	recGet := httptest.NewRecorder()
	router.handler.ServeHTTP(recGet, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if recGet.Code != http.StatusOK {
		t.Fatalf("get status = %d", recGet.Code)
	}
	var payload appsettings.PublicSettingsResponse
	if err := json.Unmarshal(recGet.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if payload.Instance.Name != "Updated Appliance" {
		t.Fatalf("instance.name = %q", payload.Instance.Name)
	}
}

func TestUpdateInstanceDisplayNameUnsupportedMethods(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, instanceSettingsPath, nil)
			req.RemoteAddr = "127.0.0.1:8080"
			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 405", rec.Code)
			}
		})
	}
}

type failingInstanceDisplayNameStore struct{}

func (f failingInstanceDisplayNameStore) UpdateInstanceDisplayNameWithAudit(
	context.Context,
	string,
	domainidentity.UserID,
	string,
) (bool, error) {
	return false, errors.New("persist failed")
}

func TestUpdateInstanceDisplayNamePersistenceFailure(t *testing.T) {
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
	login := handlers.NewLogin(handlers.LoginDeps{Service: loginService, CookiePolicy: cookie.NewPolicy(false)})
	logoutRepo := sqlite.NewLogoutRepository(store.DB())
	logout := handlers.NewLogout(handlers.LogoutDeps{Service: appidentity.NewLogoutService(logoutRepo), CookiePolicy: cookie.NewPolicy(false)})
	settingsRepo := sqlite.NewSettingsRepository(store.DB())
	settingsLoader := appsettings.NewPublicSettingsLoader(settingsRepo, settingsTestVersion, settingsTestEnvironment)
	settings := handlers.NewSettings(handlers.SettingsDeps{Loader: settingsLoader})
	updateInstance := handlers.NewUpdateInstanceSettings(handlers.UpdateInstanceSettingsDeps{
		Service: appsettings.NewUpdateInstanceDisplayNameService(failingInstanceDisplayNameStore{}),
	})
	resolver := appidentity.NewSessionResolver(sqlite.NewSessionAuthRepository(store.DB()))
	cfg := config.Config{Env: "development", ReadTimeout: 15 * time.Second}
	handler := httpapi.NewRouter(cfg, slog.Default(), health.NewReadiness(store), bootstrap, login, logout, nil, settings, updateInstance, &httpapi.SessionAuth{
		Resolver:   resolver,
		Authorizer: ports.NewRBACAuthorizer(),
	})

	bootstrapOwner(t, identityRouter{handler: handler, store: store})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: handler})
	hostnameBefore := getHostname(t, store.DB())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, instancePATCH(`{"display_name":"Fail Name"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if getHostname(t, store.DB()) != hostnameBefore {
		t.Fatal("hostname must remain unchanged on failure")
	}
}

func TestInstanceDisplayNameChainResolverRunsOnce(t *testing.T) {
	store, err := sqlite.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	inner := sqlite.NewSessionAuthRepository(store.DB())
	counter := &countingSessionAuthStore{inner: inner}
	router := newSettingsRouter(t, settingsRouterOpts{
		store:    store,
		resolver: appidentity.NewSessionResolver(counter),
	})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"Once Only"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if counter.calls.Load() != 1 {
		t.Fatalf("resolver calls = %d, want 1", counter.calls.Load())
	}
}

func TestInstanceDisplayNameFullChain(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	recBootstrap := httptest.NewRecorder()
	router.handler.ServeHTTP(recBootstrap, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if recBootstrap.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d", recBootstrap.Code)
	}

	recLogin := httptest.NewRecorder()
	router.handler.ServeHTTP(recLogin, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login status = %d", recLogin.Code)
	}
	sessionCookie := findCookie(recLogin.Result().Cookies(), cookie.SessionCookieName)
	var loginBody map[string]string
	if err := json.Unmarshal(recLogin.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	csrfToken := loginBody["csrf_token"]

	recGet1 := httptest.NewRecorder()
	router.handler.ServeHTTP(recGet1, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if recGet1.Code != http.StatusOK {
		t.Fatalf("initial get status = %d", recGet1.Code)
	}

	recPatch := httptest.NewRecorder()
	router.handler.ServeHTTP(recPatch, instancePATCH(`{"display_name":"Chain Appliance"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if recPatch.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", recPatch.Code, recPatch.Body.String())
	}

	recGet2 := httptest.NewRecorder()
	router.handler.ServeHTTP(recGet2, settingsGET([]*http.Cookie{sessionCookie}, ""))
	var settingsBody appsettings.PublicSettingsResponse
	if err := json.Unmarshal(recGet2.Body.Bytes(), &settingsBody); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if settingsBody.Instance.Name != "Chain Appliance" {
		t.Fatalf("instance.name = %q", settingsBody.Instance.Name)
	}

	if countInstanceDisplayNameAudits(t, router.audit) != 1 {
		t.Fatal("expected one success audit after change")
	}

	recNoOp := httptest.NewRecorder()
	router.handler.ServeHTTP(recNoOp, instancePATCH(`{"display_name":"Chain Appliance"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if recNoOp.Code != http.StatusOK {
		t.Fatalf("noop patch status = %d", recNoOp.Code)
	}
	if countInstanceDisplayNameAudits(t, router.audit) != 1 {
		t.Fatal("no-op must not add second audit")
	}

	recLogout := httptest.NewRecorder()
	router.handler.ServeHTTP(recLogout, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if recLogout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", recLogout.Code)
	}

	recPatchAfterLogout := httptest.NewRecorder()
	router.handler.ServeHTTP(recPatchAfterLogout, instancePATCH(`{"display_name":"After Logout"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if recPatchAfterLogout.Code != http.StatusUnauthorized {
		t.Fatalf("patch after logout status = %d", recPatchAfterLogout.Code)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range events {
		if event.EventType == appsettings.AuditEventInstanceDisplayNameUpdated {
			if event.MetadataJSON != "{}" {
				t.Fatalf("audit metadata = %q", event.MetadataJSON)
			}
			if strings.Contains(event.MetadataJSON, "Chain Appliance") {
				t.Fatal("audit leaked display name")
			}
		}
	}
}

func TestUpdateInstanceDisplayNameOversizedBody(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	oversized := `{"display_name":"` + strings.Repeat("a", 5*1024) + `"}`
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(oversized, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateInstanceDisplayNameUnicodeLength(t *testing.T) {
	name := strings.Repeat("é", setting.MaxInstanceDisplayNameRunes+1)
	if utf8.RuneCountInString(strings.TrimSpace(name)) != setting.MaxInstanceDisplayNameRunes+1 {
		t.Fatal("test setup invalid")
	}
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"display_name": name})
	router.handler.ServeHTTP(rec, instancePATCH(string(body), []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateInstanceDisplayNameAuditMetadataSafe(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"Audit Safe"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	events, err := router.audit.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range events {
		if event.EventType != appsettings.AuditEventInstanceDisplayNameUpdated {
			continue
		}
		if event.MetadataJSON != "{}" {
			t.Fatalf("metadata = %q", event.MetadataJSON)
		}
		if strings.Contains(strings.ToLower(event.MetadataJSON+event.EventType), "audit safe") {
			t.Fatal("audit leaked display name")
		}
	}
}

func TestUpdateInstanceDisplayNameDenyAuthorizerFailsClosed(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{authorizer: denyAuthorizer{}})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, instancePATCH(`{"display_name":"Denied"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateInstanceDisplayNameAuthFailuresPatchMatrix(t *testing.T) {
	const malformedBody = `{"display_name":`
	cases := []struct {
		name     string
		setup    func(t *testing.T, router settingsRouter) []*http.Cookie
		csrf     string
		wantCode int
	}{
		{
			name: "missing session cookie",
			setup: func(t *testing.T, router settingsRouter) []*http.Cookie {
				t.Helper()
				return nil
			},
			csrf:     "csrf-any",
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "unknown session token",
			setup: func(t *testing.T, router settingsRouter) []*http.Cookie {
				t.Helper()
				return []*http.Cookie{{Name: cookie.SessionCookieName, Value: "unknown-session-token"}}
			},
			csrf:     "csrf-any",
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "revoked session",
			setup: func(t *testing.T, router settingsRouter) []*http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "patch-revoked-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET revoked_at = ? WHERE id = ?`,
					"2026-07-13T12:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("revoke session: %v", err)
				}
				return []*http.Cookie{sessionCookie}
			},
			csrf:     "csrf-any",
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "idle expired session",
			setup: func(t *testing.T, router settingsRouter) []*http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "patch-idle-expired-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET idle_expires_at = ? WHERE id = ?`,
					"2000-01-01T00:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("idle expire session: %v", err)
				}
				return []*http.Cookie{sessionCookie}
			},
			csrf:     "csrf-any",
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "absolute expired session",
			setup: func(t *testing.T, router settingsRouter) []*http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "patch-absolute-expired-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET expires_at = ? WHERE id = ?`,
					"2000-01-01T00:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("absolute expire session: %v", err)
				}
				return []*http.Cookie{sessionCookie}
			},
			csrf:     "csrf-any",
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "disabled owner user",
			setup: func(t *testing.T, router settingsRouter) []*http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "patch-disabled-user-token", appidentity.UserStatusDisabled)
				return []*http.Cookie{sessionCookie}
			},
			csrf:     "csrf-any",
			wantCode: http.StatusUnauthorized,
		},
	}

	var referenceBody string
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := newSettingsRouter(t, settingsRouterOpts{})
			cookies := tc.setup(t, router)
			hostnameBefore := getHostname(t, router.store.DB())
			auditBefore := countInstanceDisplayNameAudits(t, router.audit)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, instancePATCH(malformedBody, cookies, tc.csrf))
			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
				t.Fatalf("body = %s", rec.Body.String())
			}
			normalized := normalizeSettingsErrorBody(rec.Body.Bytes())
			if i == 0 {
				referenceBody = normalized
			} else if normalized != referenceBody {
				t.Fatalf("unauthorized response differed from canonical 401 shape: %s", rec.Body.String())
			}
			assertNoSetCookie(t, rec)
			if getHostname(t, router.store.DB()) != hostnameBefore {
				t.Fatal("hostname must not change")
			}
			if countInstanceDisplayNameAudits(t, router.audit) != auditBefore {
				t.Fatal("must not append settings update audit")
			}
		})
	}
}

func TestUpdateInstanceDisplayNameCSRFFailuresPatchMatrix(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, csrfToken := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	tokenMarker := "csrf-marker-opaque-token"
	cases := []struct {
		name    string
		request func() *http.Request
	}{
		{
			name: "missing header",
			request: func() *http.Request {
				return instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, "")
			},
		},
		{
			name: "empty header",
			request: func() *http.Request {
				req := instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, "")
				req.Header.Set("X-CSRF-Token", "")
				return req
			},
		},
		{
			name: "whitespace header",
			request: func() *http.Request {
				return instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, "   ")
			},
		},
		{
			name: "oversized header",
			request: func() *http.Request {
				return instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, strings.Repeat("a", appidentity.MaxCSRFHeaderValueLen+1))
			},
		},
		{
			name: "wrong opaque token",
			request: func() *http.Request {
				return instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, tokenMarker)
			},
		},
		{
			name: "correct token in query only",
			request: func() *http.Request {
				return instancePATCHAtPath(instanceSettingsPath+"?X-CSRF-Token="+csrfToken, `{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, "")
			},
		},
		{
			name: "correct token in json body only",
			request: func() *http.Request {
				return instancePATCH(`{"display_name":"No Change","csrf_token":"`+csrfToken+`"}`, []*http.Cookie{sessionCookie}, "")
			},
		},
		{
			name: "correct token in cookie only",
			request: func() *http.Request {
				req := instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie, &http.Cookie{Name: "X-CSRF-Token", Value: csrfToken}}, "")
				return req
			},
		},
		{
			name: "alternate header name only",
			request: func() *http.Request {
				req := instancePATCH(`{"display_name":"No Change"}`, []*http.Cookie{sessionCookie}, "")
				req.Header.Set("X-Xsrf-Token", csrfToken)
				return req
			},
		},
	}

	hostnameBefore := getHostname(t, router.store.DB())
	auditBefore := countInstanceDisplayNameAudits(t, router.audit)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := tc.request()
			router.handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
				t.Fatalf("body = %s", rec.Body.String())
			}
			if parseSettingsErrorMessage(t, rec.Body.Bytes()) != "CSRF validation failed" {
				t.Fatalf("unexpected error message: %s", rec.Body.String())
			}
			assertNoSetCookie(t, rec)
			assertResponseDoesNotContainToken(t, rec, csrfToken)
			assertResponseDoesNotContainToken(t, rec, tokenMarker)
			if getHostname(t, router.store.DB()) != hostnameBefore {
				t.Fatal("hostname must not change")
			}
			if countInstanceDisplayNameAudits(t, router.audit) != auditBefore {
				t.Fatal("must not append settings update audit")
			}
		})
	}
}

func TestUpdateInstanceDisplayNameOrderingInvalidSessionBeforeCSRF(t *testing.T) {
	const malformedBody = `{"display_name":`
	sessionCases := []struct {
		name   string
		cookie func(t *testing.T, router settingsRouter) *http.Cookie
	}{
		{
			name: "unknown session",
			cookie: func(t *testing.T, router settingsRouter) *http.Cookie {
				t.Helper()
				return &http.Cookie{Name: cookie.SessionCookieName, Value: "unknown-session-token"}
			},
		},
		{
			name: "revoked session",
			cookie: func(t *testing.T, router settingsRouter) *http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "ordering-revoked-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET revoked_at = ? WHERE id = ?`,
					"2026-07-13T12:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("revoke session: %v", err)
				}
				return sessionCookie
			},
		},
		{
			name: "idle expired session",
			cookie: func(t *testing.T, router settingsRouter) *http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "ordering-idle-expired-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET idle_expires_at = ? WHERE id = ?`,
					"2000-01-01T00:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("idle expire session: %v", err)
				}
				return sessionCookie
			},
		},
		{
			name: "absolute expired session",
			cookie: func(t *testing.T, router settingsRouter) *http.Cookie {
				t.Helper()
				sessionCookie := createSettingsSession(t, router, "ordering-absolute-expired-token", appidentity.UserStatusActive)
				_, err := router.store.DB().ExecContext(
					context.Background(),
					`UPDATE sessions SET expires_at = ? WHERE id = ?`,
					"2000-01-01T00:00:00Z",
					"settings-sess-1",
				)
				if err != nil {
					t.Fatalf("absolute expire session: %v", err)
				}
				return sessionCookie
			},
		},
		{
			name: "disabled user",
			cookie: func(t *testing.T, router settingsRouter) *http.Cookie {
				t.Helper()
				return createSettingsSession(t, router, "ordering-disabled-user-token", appidentity.UserStatusDisabled)
			},
		},
	}

	csrfCases := []struct {
		name  string
		token string
	}{
		{name: "missing csrf header", token: ""},
		{name: "wrong csrf header", token: "wrong-csrf"},
	}

	for _, sessionCase := range sessionCases {
		for _, csrfCase := range csrfCases {
			t.Run(sessionCase.name+"_"+csrfCase.name, func(t *testing.T) {
				router := newSettingsRouter(t, settingsRouterOpts{})
				sessionCookie := sessionCase.cookie(t, router)

				rec := httptest.NewRecorder()
				router.handler.ServeHTTP(rec, instancePATCH(malformedBody, []*http.Cookie{sessionCookie}, csrfCase.token))
				if rec.Code != http.StatusUnauthorized {
					t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
				}
				if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
					t.Fatalf("body = %s", rec.Body.String())
				}
				assertNoSetCookie(t, rec)
			})
		}
	}
}

func TestUpdateInstanceDisplayNameOrderingPermissionBeforeCSRF(t *testing.T) {
	const malformedBody = `{"display_name":`
	csrfCases := []struct {
		name  string
		token string
	}{
		{name: "missing csrf header", token: ""},
		{name: "wrong csrf header", token: "wrong-csrf"},
	}

	for _, csrfCase := range csrfCases {
		t.Run(csrfCase.name, func(t *testing.T) {
			router := newSettingsRouter(t, settingsRouterOpts{})
			sessionCookie := createRoleSettingsSession(
				t,
				router,
				domainidentity.UserID("permission-order-admin"),
				"permission-order-admin",
				domainidentity.RoleAdministrator,
				"permission-order-admin-session",
				"permission-order-admin-token",
			)
			hostnameBefore := getHostname(t, router.store.DB())
			auditBefore := countInstanceDisplayNameAudits(t, router.audit)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, instancePATCH(malformedBody, []*http.Cookie{sessionCookie}, csrfCase.token))
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
				t.Fatalf("body = %s", rec.Body.String())
			}
			if parseSettingsErrorMessage(t, rec.Body.Bytes()) != "Permission denied" {
				t.Fatalf("unexpected error message: %s", rec.Body.String())
			}
			assertNoSetCookie(t, rec)
			if getHostname(t, router.store.DB()) != hostnameBefore {
				t.Fatal("hostname must not change")
			}
			if countInstanceDisplayNameAudits(t, router.audit) != auditBefore {
				t.Fatal("must not append settings update audit")
			}
		})
	}
}

func TestUpdateInstanceDisplayNameOrderingValidOwnerInvalidCSRFGivesCSRF403(t *testing.T) {
	const malformedBody = `{"display_name":`
	router := newSettingsRouter(t, settingsRouterOpts{})
	bootstrapOwner(t, identityRouter{handler: router.handler})
	sessionCookie, _ := loginAndGetCredentials(t, identityRouter{handler: router.handler})

	csrfCases := []struct {
		name  string
		token string
	}{
		{name: "missing csrf header", token: ""},
		{name: "wrong csrf header", token: "wrong-csrf"},
	}

	hostnameBefore := getHostname(t, router.store.DB())
	auditBefore := countInstanceDisplayNameAudits(t, router.audit)

	for _, csrfCase := range csrfCases {
		t.Run(csrfCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, instancePATCH(malformedBody, []*http.Cookie{sessionCookie}, csrfCase.token))
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
				t.Fatalf("body = %s", rec.Body.String())
			}
			if parseSettingsErrorMessage(t, rec.Body.Bytes()) != "CSRF validation failed" {
				t.Fatalf("unexpected error message: %s", rec.Body.String())
			}
			assertNoSetCookie(t, rec)
			if getHostname(t, router.store.DB()) != hostnameBefore {
				t.Fatal("hostname must not change")
			}
			if countInstanceDisplayNameAudits(t, router.audit) != auditBefore {
				t.Fatal("must not append settings update audit")
			}
		})
	}
}
