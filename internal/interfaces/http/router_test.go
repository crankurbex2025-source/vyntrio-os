package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installmediapublic"
)

type stubDB struct{}

func (stubDB) Ping(_ context.Context) error { return nil }

func testRouter(t *testing.T, opts ...RouterOption) http.Handler {
	t.Helper()
	cfg := config.Config{
		Version:     "0.2.0-dev",
		BuildCommit: "test",
		ReadTimeout: 15 * time.Second,
	}
	logger := slog.Default()
	readiness := health.NewReadiness(stubDB{})
	return NewRouter(cfg, logger, readiness, nil, nil, nil, nil, nil, nil, nil, nil, opts...)
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

func TestRouterPublicInstallMediaNotBuiltWithoutStaging(t *testing.T) {
	r := testRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/public/install-media", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var got installmediapublic.Metadata
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if got.PublicationStatus != installmediapublic.PublicationNotBuilt {
		t.Fatalf("status = %q", got.PublicationStatus)
	}
}

func TestRouterPublicInstallMediaReleaseDownload(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, installmediapublic.ArtifactName)
	if err := os.WriteFile(artifactPath, []byte("stage-test-image"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	public := installmediapublic.Metadata{
		PublicationStatus: installmediapublic.PublicationLocalStaging,
		GeneratedAt:       "2026-07-16T12:00:00Z",
		Release: installmediapublic.ReleaseLine{
			Version: "0.2.0-dev",
			Channel: "development",
		},
		PrimaryArtifact: installmediapublic.PrimaryArtifact{
			Name:             installmediapublic.ArtifactName,
			Format:           "raw_gpt_hybrid_disk",
			FirmwareBootMode: "bios+uefi",
			DownloadPath:     "/release/" + installmediapublic.ArtifactName,
		},
		BuildTarget: "make install-media",
		StageTarget: "make release-install-media-stage",
	}
	data, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, installmediapublic.PublicMetadataName), data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	r := testRouter(t, WithReleaseStaging(dir))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/public/install-media", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metadata status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/release/"+installmediapublic.ArtifactName, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("download status = %d", rec.Code)
	}
	if rec.Body.String() != "stage-test-image" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
