package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

type stubDB struct{}

func (stubDB) Ping(_ context.Context) error { return nil }

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Config{
		Version:     "0.2.0-dev",
		BuildCommit: "test",
		ReadTimeout: 15 * time.Second,
	}
	logger := slog.Default()
	readiness := health.NewReadiness(stubDB{})
	return NewRouter(cfg, logger, readiness)
}

func TestRouterHealthz(t *testing.T) {
	r := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get(middleware.RequestIDHeader) == "" {
		t.Error("missing X-Request-ID header")
	}
}

func TestRouterReadyz(t *testing.T) {
	r := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterVersion(t *testing.T) {
	r := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["version"] != "0.2.0-dev" {
		t.Errorf("version = %q", body["version"])
	}
}

func TestRouterNotFound(t *testing.T) {
	r := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestRouterRequestIDPreserved(t *testing.T) {
	r := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(middleware.RequestIDHeader, "client-id-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.RequestIDHeader); got != "client-id-123" {
		t.Errorf("X-Request-ID = %q, want client-id-123", got)
	}
}
