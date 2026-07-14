package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/ui"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

const (
	uiTestHTMLCache  = "no-cache"
	uiTestAssetCache = "public, max-age=31536000, immutable"
	uiTestCSP        = "default-src 'self'; script-src 'self'; style-src 'self'; " +
		"img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; " +
		"form-action 'self'; frame-ancestors 'none'"
)

func testUIRouter(t *testing.T) http.Handler {
	t.Helper()
	uiHandler, err := ui.NewHandler()
	if err != nil {
		t.Fatalf("ui.NewHandler() error: %v", err)
	}
	cfg := config.Config{
		Version:     "0.2.0-dev",
		BuildCommit: "test",
		ReadTimeout: 15 * time.Second,
	}
	readiness := health.NewReadiness(stubDB{})
	return NewRouter(cfg, slog.Default(), readiness, nil, nil, nil, nil, nil, nil, nil, WithUI(uiHandler))
}

func assertUIHTMLHeaders(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	header := rec.Header()
	if got := header.Get("Cache-Control"); got != uiTestHTMLCache {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := header.Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := header.Get("Content-Security-Policy"); got != uiTestCSP {
		t.Fatalf("CSP = %q", got)
	}
	if got := header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := header.Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q", got)
	}
}

func TestUIRouterRootServesEmbeddedIndex(t *testing.T) {
	r := testUIRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertUIHTMLHeaders(t, rec)
	if !strings.Contains(rec.Body.String(), `<div id="root"></div>`) {
		t.Fatal("root must serve embedded index.html")
	}
}

func TestUIRouterSPAFallbackServesIndex(t *testing.T) {
	r := testUIRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/future-route", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertUIHTMLHeaders(t, rec)
	if !strings.Contains(rec.Body.String(), `<div id="root"></div>`) {
		t.Fatal("SPA fallback must serve embedded index.html")
	}
}

func TestUIRouterHEADRootAndFallbackHaveNoBody(t *testing.T) {
	r := testUIRouter(t)
	for _, target := range []string{"/", "/future-route"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, target, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("HEAD %s status = %d", target, rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("HEAD %s body length = %d, want 0", target, rec.Body.Len())
		}
		assertUIHTMLHeaders(t, rec)
	}
}

func TestUIRouterMutatingMethodsDoNotServeIndex(t *testing.T) {
	r := testUIRouter(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		for _, target := range []string{"/", "/future-route"} {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
			if rec.Code != http.StatusNotFound {
				t.Fatalf("%s %s status = %d, want 404", method, target, rec.Code)
			}
			if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("%s %s must keep canonical JSON 404, got %q", method, target, rec.Header().Get("Content-Type"))
			}
			if strings.Contains(rec.Body.String(), "<html") {
				t.Fatalf("%s %s must not serve HTML", method, target)
			}
		}
	}
}

func TestUIRouterUnknownAPIPathKeepsCanonicalJSON404(t *testing.T) {
	r := testUIRouter(t)
	methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, method := range methods {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(method, "/api/v1/does-not-exist", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", method, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("%s Content-Type = %q, want application/json", method, got)
		}
		if rec.Header().Get("Content-Security-Policy") != "" {
			t.Fatalf("%s API 404 must not carry UI CSP", method)
		}
		if rec.Header().Get("Cache-Control") == uiTestHTMLCache {
			t.Fatalf("%s API 404 must not carry UI cache policy", method)
		}
		if method == http.MethodHead {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("%s body is not JSON: %v", method, err)
		}
		errObj, ok := payload["error"].(map[string]any)
		if !ok {
			t.Fatalf("%s missing error envelope: %v", method, payload)
		}
		if errObj["code"] != "NOT_FOUND" {
			t.Fatalf("%s code = %v", method, errObj["code"])
		}
	}
}

func TestUIRouterReservedPrefixRootsAreNotSPAFallback(t *testing.T) {
	r := testUIRouter(t)
	for _, target := range []string{"/api", "/api/", "/api/v2/settings", "/assets"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET %s status = %d, want 404", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("GET %s must not serve index.html", target)
		}
	}
}

func TestUIRouterHealthProbesUnchanged(t *testing.T) {
	r := testUIRouter(t)
	for _, target := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("GET %s must not serve HTML", target)
		}
		if rec.Header().Get("Content-Security-Policy") != "" {
			t.Fatalf("GET %s must not carry UI CSP", target)
		}
	}
}

func TestUIRouterVersionEndpointUnchanged(t *testing.T) {
	r := testUIRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["version"] != "0.2.0-dev" {
		t.Fatalf("version = %q", body["version"])
	}
}

func TestUIRouterServesAssetsWithImmutableCache(t *testing.T) {
	r := testUIRouter(t)

	recIndex := httptest.NewRecorder()
	r.ServeHTTP(recIndex, httptest.NewRequest(http.MethodGet, "/", nil))
	assetPath := extractFirstAssetPath(t, recIndex.Body.String())

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, assetPath, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d", assetPath, rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != uiTestAssetCache {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if rec.Header().Get("Content-Security-Policy") != "" {
		t.Fatal("asset response must not carry UI CSP header")
	}

	recHead := httptest.NewRecorder()
	r.ServeHTTP(recHead, httptest.NewRequest(http.MethodHead, assetPath, nil))
	if recHead.Code != http.StatusOK {
		t.Fatalf("HEAD %s status = %d", assetPath, recHead.Code)
	}
	if recHead.Body.Len() != 0 {
		t.Fatalf("HEAD %s body length = %d, want 0", assetPath, recHead.Body.Len())
	}
}

func TestUIRouterMissingAssetIsNotIndexAndNotImmutable(t *testing.T) {
	r := testUIRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/not-found.js", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatal("missing asset must not serve index.html")
	}
	if rec.Header().Get("Cache-Control") == uiTestAssetCache {
		t.Fatal("missing asset must not receive immutable cache")
	}
}

func TestUIRouterWithoutUIKeepsExistingNotFound(t *testing.T) {
	r := testRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatal("router without UI must not serve HTML")
	}
}

// extractFirstAssetPath pulls a /assets/... reference out of served HTML so
// the test follows the real emitted asset names.
func extractFirstAssetPath(t *testing.T, html string) string {
	t.Helper()
	start := strings.Index(html, `src="/assets/`)
	if start == -1 {
		t.Fatal("index.html does not reference /assets/ script")
	}
	start += len(`src="`)
	end := strings.Index(html[start:], `"`)
	if end == -1 {
		t.Fatal("unterminated asset src attribute")
	}
	return html[start : start+end]
}
