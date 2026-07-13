package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
)

type countingSessionResolver struct {
	inner *appidentity.SessionResolver
	calls atomic.Int32
}

func (c *countingSessionResolver) Resolve(ctx context.Context, rawSessionToken string) (appidentity.ResolvedSession, bool, error) {
	c.calls.Add(1)
	return c.inner.Resolve(ctx, rawSessionToken)
}

func TestRequireCSRFValidHeaderReachesHandler(t *testing.T) {
	env := newAuthTestEnv(t)
	rawSession := "session-token-for-csrf-test"
	rawCSRF := "csrf-token-for-csrf-test"
	env.createActiveOwnerSession(t, rawSession, appidentity.HashRawToken(rawCSRF))

	called := false
	handler := middleware.OptionalAuthentication(env.resolver)(
		middleware.RequireAuthentication(
			middleware.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: rawSession})
	req.Header.Set("X-CSRF-Token", rawCSRF)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("status = %d called = %v", rec.Code, called)
	}
}

func TestRequireCSRFRejectsMissingAndInvalid(t *testing.T) {
	env := newAuthTestEnv(t)
	rawSession := "session-token-csrf-reject"
	rawCSRF := "valid-csrf-token-value"
	env.createActiveOwnerSession(t, rawSession, appidentity.HashRawToken(rawCSRF))

	called := false
	handler := middleware.OptionalAuthentication(env.resolver)(
		middleware.RequireAuthentication(
			middleware.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})),
		),
	)

	cases := []struct {
		name   string
		header string
	}{
		{name: "missing", header: ""},
		{name: "empty", header: ""},
		{name: "whitespace", header: "   "},
		{name: "wrong", header: "wrong-csrf-token"},
		{name: "oversized", header: strings.Repeat("a", appidentity.MaxCSRFHeaderValueLen+1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called = false
			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: rawSession})
			if tc.header != "" {
				req.Header.Set("X-CSRF-Token", tc.header)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", rec.Code)
			}
			if called {
				t.Fatal("handler must not run on CSRF failure")
			}
			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error: %v", err)
			}
			errObj, _ := payload["error"].(map[string]any)
			if errObj["code"] != "FORBIDDEN" || errObj["message"] != "CSRF validation failed" {
				t.Fatalf("error = %v", errObj)
			}
		})
	}
}

func TestRequireCSRFHeaderOnlyNotQueryOrCookie(t *testing.T) {
	env := newAuthTestEnv(t)
	rawSession := "session-token-header-only"
	rawCSRF := "header-only-csrf-token"
	env.createActiveOwnerSession(t, rawSession, appidentity.HashRawToken(rawCSRF))

	called := false
	handler := middleware.OptionalAuthentication(env.resolver)(
		middleware.RequireAuthentication(
			middleware.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/logout?X-CSRF-Token="+rawCSRF, strings.NewReader("csrf_token="+rawCSRF))
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: rawSession})
	req.AddCookie(&http.Cookie{Name: "X-CSRF-Token", Value: rawCSRF})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Xsrf-Token", rawCSRF)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || called {
		t.Fatalf("status = %d called = %v", rec.Code, called)
	}
}

func TestRequireCSRFUnauthorizedBeforeEvaluation(t *testing.T) {
	env := newAuthTestEnv(t)
	called := false
	handler := middleware.OptionalAuthentication(env.resolver)(
		middleware.RequireAuthentication(
			middleware.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("X-CSRF-Token", "any")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || called {
		t.Fatalf("status = %d called = %v", rec.Code, called)
	}
}

func TestAuthCSRFChainResolvesSessionOnce(t *testing.T) {
	env := newAuthTestEnv(t)
	rawSession := "session-once-resolve"
	rawCSRF := "csrf-once-resolve"
	env.createActiveOwnerSession(t, rawSession, appidentity.HashRawToken(rawCSRF))

	counter := &countingSessionResolver{inner: env.resolver}
	called := false
	handler := middleware.OptionalAuthentication(counter)(
		middleware.RequireAuthentication(
			middleware.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: cookie.SessionCookieName, Value: rawSession})
	req.Header.Set("X-CSRF-Token", rawCSRF)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("status = %d called = %v", rec.Code, called)
	}
	if counter.calls.Load() != 1 {
		t.Fatalf("resolver calls = %d, want 1", counter.calls.Load())
	}
}