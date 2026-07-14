package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appoverview "github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

const (
	overviewPath            = "/api/v1/overview"
	overviewTestVersion     = "0.2.0-test"
	overviewTestEnvironment = "development"
)

func overviewGET(cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, overviewPath, nil)
	req.RemoteAddr = "127.0.0.1:8080"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func TestOverviewUnauthenticatedReturns401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET(nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestOverviewMissingPermissionReturns403(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		authorizer: denyAuthorizer{},
	})

	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestOverviewRealRBACAuthorizationMatrix(t *testing.T) {
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
			userID:     domainidentity.UserID("overview-owner"),
			username:   "overview-owner",
			role:       domainidentity.RoleOwner,
			sessionID:  "overview-owner-sess",
			rawToken:   "overview-owner-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "operator",
			userID:     domainidentity.UserID("overview-operator"),
			username:   "overview-operator",
			role:       domainidentity.RoleOperator,
			sessionID:  "overview-operator-sess",
			rawToken:   "overview-operator-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "read_only",
			userID:     domainidentity.UserID("overview-read-only"),
			username:   "overview-read-only",
			role:       domainidentity.RoleReadOnly,
			sessionID:  "overview-read-only-sess",
			rawToken:   "overview-read-only-token",
			wantStatus: http.StatusOK,
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
			router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
			if tc.wantStatus != http.StatusOK {
				return
			}

			assertOverviewResponseShape(t, rec)
			assertOverviewCacheControlNoStore(t, rec)
			assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
		})
	}
}

func TestOverviewResponseShapeOwnerSession(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Instance.Name == "" {
		t.Fatal("expected instance name")
	}
	if got.Instance.Version != settingsTestVersion {
		t.Fatalf("version = %q, want %q", got.Instance.Version, settingsTestVersion)
	}
	if got.Instance.Commit != "test-commit" {
		t.Fatalf("commit = %q, want test-commit", got.Instance.Commit)
	}
	if got.API.Environment != settingsTestEnvironment {
		t.Fatalf("environment = %q, want %q", got.API.Environment, settingsTestEnvironment)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q, want running", got.Service.Status)
	}
	if got.Readiness.Status != "ready" {
		t.Fatalf("readiness.status = %q, want ready", got.Readiness.Status)
	}
	if got.Readiness.Database != "ok" {
		t.Fatalf("readiness.database = %q, want ok", got.Readiness.Database)
	}
	if got.CollectedAt == "" {
		t.Fatal("expected collected_at")
	}
	if _, err := time.Parse(time.RFC3339Nano, got.CollectedAt); err != nil {
		t.Fatalf("collected_at parse error: %v", err)
	}

	assertOverviewCacheControlNoStore(t, rec)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}

func TestOverviewDatabaseFailureReturns200NotReady(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		readinessDB: failingDBChecker{},
	})

	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}

	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Readiness.Status != "not_ready" {
		t.Fatalf("readiness.status = %q, want not_ready", got.Readiness.Status)
	}
	if got.Readiness.Database != "error" {
		t.Fatalf("readiness.database = %q, want error", got.Readiness.Database)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q, want running", got.Service.Status)
	}
}

func TestOverviewPreservesHealthEndpointBehavior(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("readyz status = %d, want 200", rec.Code)
	}
}

type failingDBChecker struct{}

func (failingDBChecker) Ping(context.Context) error {
	return context.Canceled
}

func decodeOverviewResponse(t *testing.T, body []byte) appoverview.Response {
	t.Helper()
	var got appoverview.Response
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v body=%s", err, body)
	}
	return got
}

func assertOverviewResponseShape(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Instance.Name == "" || got.Instance.Version == "" || got.Instance.Commit == "" {
		t.Fatalf("instance = %+v", got.Instance)
	}
	if got.API.Environment == "" {
		t.Fatalf("api = %+v", got.API)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service = %+v", got.Service)
	}
	if got.Readiness.Status == "" || got.Readiness.Database == "" {
		t.Fatalf("readiness = %+v", got.Readiness)
	}
	if got.CollectedAt == "" {
		t.Fatal("missing collected_at")
	}
}

func assertOverviewCacheControlNoStore(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func assertNoSensitiveOverviewFields(t *testing.T, body []byte) {
	t.Helper()
	lower := strings.ToLower(string(body))
	for _, forbidden := range []string{
		"password", "token", "hash", "csrf", "session", "userid", "user_id",
		"role", "principal", "datadir", "data_dir", "path", "port",
		"bind", "timezone", "127.0.0.1", "sqlite", "audit", "config.toml",
		"/var/lib", "/etc/vyntrio",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("response leaked forbidden substring %q: %s", forbidden, body)
		}
	}
}
