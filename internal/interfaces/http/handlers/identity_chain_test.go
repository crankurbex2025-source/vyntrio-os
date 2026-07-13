package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
)

func TestIdentityChainBootstrapLoginSettingsLogout(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	ctx := context.Background()

	recBootstrap := httptest.NewRecorder()
	router.handler.ServeHTTP(recBootstrap, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if recBootstrap.Code != http.StatusCreated {
		t.Fatalf("bootstrap status = %d, want 201", recBootstrap.Code)
	}

	recLogin := httptest.NewRecorder()
	router.handler.ServeHTTP(recLogin, loginPOST(`{"username":"owner","password":"`+testLoginPassword+`"}`, nil))
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", recLogin.Code, recLogin.Body.String())
	}

	loginCookies := recLogin.Result().Cookies()
	if len(loginCookies) != 1 {
		t.Fatalf("login Set-Cookie count = %d, want 1", len(loginCookies))
	}
	sessionCookie := findCookie(loginCookies, "vyntrio_session")
	if sessionCookie == nil {
		t.Fatal("missing vyntrio_session cookie")
	}
	if sessionCookie.HttpOnly != true {
		t.Fatal("session cookie must be HttpOnly")
	}

	var loginBody map[string]string
	if err := json.Unmarshal(recLogin.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if len(loginBody) != 1 {
		t.Fatalf("login body keys = %v, want csrf_token only", loginBody)
	}
	csrfToken, ok := loginBody["csrf_token"]
	if !ok || csrfToken == "" {
		t.Fatalf("login body = %v", loginBody)
	}
	if csrfToken == sessionCookie.Value {
		t.Fatal("csrf_token must not equal session cookie")
	}

	recSettings := httptest.NewRecorder()
	router.handler.ServeHTTP(recSettings, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if recSettings.Code != http.StatusOK {
		t.Fatalf("settings status = %d body=%s", recSettings.Code, recSettings.Body.String())
	}
	assertNoSensitiveSettingsFields(t, recSettings.Body.Bytes())
	if strings.Contains(recSettings.Body.String(), sessionCookie.Value) || strings.Contains(recSettings.Body.String(), csrfToken) {
		t.Fatal("settings response leaked session or CSRF material")
	}

	recLogout := httptest.NewRecorder()
	router.handler.ServeHTTP(recLogout, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if recLogout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d body=%s", recLogout.Code, recLogout.Body.String())
	}
	cleared := findCookie(recLogout.Result().Cookies(), "vyntrio_session")
	if cleared == nil || cleared.MaxAge != -1 || cleared.Value != "" {
		t.Fatalf("logout cookie clear = %#v", cleared)
	}
	if strings.Contains(recLogout.Body.String(), sessionCookie.Value) || strings.Contains(recLogout.Body.String(), csrfToken) {
		t.Fatal("logout response leaked secrets")
	}

	recSettingsAfter := httptest.NewRecorder()
	router.handler.ServeHTTP(recSettingsAfter, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if recSettingsAfter.Code != http.StatusUnauthorized {
		t.Fatalf("settings after logout status = %d, want 401", recSettingsAfter.Code)
	}
	if parseSettingsErrorCode(t, recSettingsAfter.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("settings after logout error = %s", recSettingsAfter.Body.String())
	}
	assertNoSensitiveSettingsFields(t, recSettingsAfter.Body.Bytes())

	events, err := router.audit.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	for _, event := range events {
		if strings.Contains(event.MetadataJSON, sessionCookie.Value) ||
			strings.Contains(event.MetadataJSON, csrfToken) {
			t.Fatalf("audit metadata leaked secret: event=%s metadata=%s", event.EventType, event.MetadataJSON)
		}
	}

	rows, err := router.store.DB().QueryContext(ctx, `SELECT session_token_hash, csrf_token_hash FROM sessions`)
	if err != nil {
		t.Fatalf("query sessions: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sessionHash, csrfHash string
		if err := rows.Scan(&sessionHash, &csrfHash); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if sessionHash == sessionCookie.Value || csrfHash == csrfToken {
			t.Fatal("database must store hashes only, not raw tokens")
		}
	}
}
