package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/ui"
)

func newUISettingsRouter(t *testing.T) settingsRouter {
	t.Helper()
	uiHandler, err := ui.NewHandler()
	if err != nil {
		t.Fatalf("ui.NewHandler() error: %v", err)
	}
	return newSettingsRouter(t, settingsRouterOpts{
		routerOpts: []httpapi.RouterOption{httpapi.WithUI(uiHandler)},
	})
}

func TestUIEnabledSettingsGETKeepsCanonicalAuthBehavior(t *testing.T) {
	router := newUISettingsRouter(t)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, settingsGET(nil, ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if rec.Header().Get("Content-Security-Policy") != "" {
		t.Fatal("API response must not carry UI CSP")
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatal("API 401 must not return HTML")
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestUIEnabledFullIdentitySettingsChainStillWorks(t *testing.T) {
	router := newUISettingsRouter(t)

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
	sessionCookie := recLogin.Result().Cookies()[0]
	var loginBody map[string]string
	if err := json.Unmarshal(recLogin.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	csrfToken := loginBody["csrf_token"]

	recGet := httptest.NewRecorder()
	router.handler.ServeHTTP(recGet, settingsGET([]*http.Cookie{sessionCookie}, ""))
	if recGet.Code != http.StatusOK {
		t.Fatalf("settings status = %d", recGet.Code)
	}

	recPatch := httptest.NewRecorder()
	router.handler.ServeHTTP(recPatch, instancePATCH(`{"display_name":"UI Enabled Appliance"}`, []*http.Cookie{sessionCookie}, csrfToken))
	if recPatch.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", recPatch.Code, recPatch.Body.String())
	}

	recLogout := httptest.NewRecorder()
	router.handler.ServeHTTP(recLogout, logoutPOST([]*http.Cookie{sessionCookie}, csrfToken))
	if recLogout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", recLogout.Code)
	}
}

func TestUIEnabledUnknownAPIRouteKeepsJSON404(t *testing.T) {
	router := newUISettingsRouter(t)
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete} {
		req := httptest.NewRequest(method, "/api/v1/does-not-exist", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		router.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", method, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("%s Content-Type = %q", method, got)
		}
		if parseSettingsErrorCode(t, rec.Body.Bytes()) != "NOT_FOUND" {
			t.Fatalf("%s body = %s", method, rec.Body.String())
		}
	}
}

func TestUIEnabledRootServesHTMLAndProbesUnchanged(t *testing.T) {
	router := newUISettingsRouter(t)

	recRoot := httptest.NewRecorder()
	rootReq := httptest.NewRequest(http.MethodGet, "/", nil)
	rootReq.RemoteAddr = "127.0.0.1:8080"
	router.handler.ServeHTTP(recRoot, rootReq)
	if recRoot.Code != http.StatusOK {
		t.Fatalf("root status = %d", recRoot.Code)
	}
	if got := recRoot.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("root Content-Type = %q", got)
	}

	for _, target := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		req.RemoteAddr = "127.0.0.1:8080"
		router.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("%s must not serve HTML", target)
		}
	}
}
