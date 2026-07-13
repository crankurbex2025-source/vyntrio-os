package ui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	wantHTMLCache  = "no-cache"
	wantAssetCache = "public, max-age=31536000, immutable"
	wantCSP        = "default-src 'self'; script-src 'self'; style-src 'self'; " +
		"img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; " +
		"form-action 'self'; frame-ancestors 'none'"
)

func embeddedAssetNames(t *testing.T) (js string, css string) {
	t.Helper()
	entries, err := fs.ReadDir(distFS, "dist/assets")
	if err != nil {
		t.Fatalf("read embedded assets: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		switch {
		case strings.HasSuffix(name, ".js"):
			js = name
		case strings.HasSuffix(name, ".css"):
			css = name
		}
	}
	if js == "" || css == "" {
		t.Fatalf("embedded dist/assets missing js or css: js=%q css=%q", js, css)
	}
	return js, css
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error: %v", err)
	}
	return h
}

func TestNewHandlerValidatesEmbeddedAssets(t *testing.T) {
	h := newTestHandler(t)
	if len(h.index) == 0 {
		t.Fatal("index must not be empty")
	}
	if !strings.Contains(string(h.index), "/assets/") {
		t.Fatal("embedded index.html must reference /assets/")
	}
}

func TestServeIndexGETHeadersAndBody(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeIndex(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	header := rec.Header()
	if got := header.Get("Cache-Control"); got != wantHTMLCache {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := header.Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := header.Get("Content-Security-Policy"); got != wantCSP {
		t.Fatalf("CSP = %q", got)
	}
	if got := header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := header.Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q", got)
	}
	if !strings.Contains(rec.Body.String(), "<div id=\"root\"></div>") {
		t.Fatal("index body must contain react root element")
	}
}

func TestServeIndexHEADHasHeadersNoBody(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeIndex(rec, httptest.NewRequest(http.MethodHead, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD body length = %d, want 0", rec.Body.Len())
	}
	if got := rec.Header().Get("Cache-Control"); got != wantHTMLCache {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got != wantCSP {
		t.Fatalf("CSP = %q", got)
	}
}

func TestServeIndexRejectsMutatingMethods(t *testing.T) {
	h := newTestHandler(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		rec := httptest.NewRecorder()
		h.ServeIndex(rec, httptest.NewRequest(method, "/", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", method, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("%s must not return HTML", method)
		}
	}
}

func TestServeAssetGETJavaScriptAndCSS(t *testing.T) {
	h := newTestHandler(t)
	js, css := embeddedAssetNames(t)

	cases := []struct {
		name        string
		contentType string
	}{
		{name: js, contentType: "text/javascript; charset=utf-8"},
		{name: css, contentType: "text/css; charset=utf-8"},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		h.ServeAsset(rec, httptest.NewRequest(http.MethodGet, "/assets/"+tc.name, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", tc.name, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != tc.contentType {
			t.Fatalf("%s Content-Type = %q, want %q", tc.name, got, tc.contentType)
		}
		if got := rec.Header().Get("Cache-Control"); got != wantAssetCache {
			t.Fatalf("%s Cache-Control = %q", tc.name, got)
		}
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("%s X-Content-Type-Options = %q", tc.name, got)
		}
		if got := rec.Header().Get("Content-Security-Policy"); got != "" {
			t.Fatalf("%s must not carry HTML CSP header, got %q", tc.name, got)
		}
		if rec.Body.Len() == 0 {
			t.Fatalf("%s body must not be empty", tc.name)
		}
		want, err := distFS.ReadFile("dist/assets/" + tc.name)
		if err != nil {
			t.Fatalf("read embedded %s: %v", tc.name, err)
		}
		if rec.Body.String() != string(want) {
			t.Fatalf("%s bytes differ from embedded file", tc.name)
		}
	}
}

func TestServeAssetHEADHasHeadersNoBody(t *testing.T) {
	h := newTestHandler(t)
	js, _ := embeddedAssetNames(t)

	rec := httptest.NewRecorder()
	h.ServeAsset(rec, httptest.NewRequest(http.MethodHead, "/assets/"+js, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD body length = %d, want 0", rec.Body.Len())
	}
	if got := rec.Header().Get("Cache-Control"); got != wantAssetCache {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestServeAssetMissingIsPlainNotFoundWithoutImmutableCache(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeAsset(rec, httptest.NewRequest(http.MethodGet, "/assets/not-found.js", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("Content-Type = %q, want text/plain", got)
	}
	if got := rec.Header().Get("Cache-Control"); got == wantAssetCache {
		t.Fatal("missing asset must not receive immutable cache")
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatal("missing asset must not return HTML")
	}
}

func TestServeAssetRejectsTraversalAndMalformedPaths(t *testing.T) {
	h := newTestHandler(t)
	paths := []string{
		"/assets/../index.html",
		"/assets/..",
		"/assets/./index.html",
		"/assets/",
		"/assets/sub/",
		"/assets/.hidden",
		"/assets/dir/.hidden",
		"/assets/..\\index.html",
		"/assets/a\\b.js",
		"/assets//double.js",
	}
	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.URL.Path = p
		rec := httptest.NewRecorder()
		h.ServeAsset(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", p, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("%s must not fall back to index.html", p)
		}
	}
}

func TestServeAssetRejectsMutatingMethods(t *testing.T) {
	h := newTestHandler(t)
	js, _ := embeddedAssetNames(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		rec := httptest.NewRecorder()
		h.ServeAsset(rec, httptest.NewRequest(method, "/assets/"+js, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", method, rec.Code)
		}
	}
}

func TestEmbeddedIndexIsCSPCompatible(t *testing.T) {
	h := newTestHandler(t)
	html := string(h.index)

	if strings.Contains(html, "<script") && !strings.Contains(html, `src="/assets/`) {
		t.Fatal("script tag without same-origin /assets/ src")
	}
	for _, forbidden := range []string{
		"<style", "javascript:", "http://", "https://", "onload=", "onclick=",
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("index.html contains CSP-incompatible content: %q", forbidden)
		}
	}
	scriptCount := strings.Count(html, "<script")
	srcCount := strings.Count(html, `<script type="module" crossorigin src="/assets/`)
	if scriptCount != srcCount {
		t.Fatalf("found %d script tags but %d same-origin module scripts", scriptCount, srcCount)
	}
	if !strings.Contains(html, `href="/assets/`) {
		t.Fatal("stylesheet link must reference same-origin /assets/")
	}
}
