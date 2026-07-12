package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
)

type pingOK struct{}

func (pingOK) Ping(_ context.Context) error { return nil }

type pingFail struct{}

func (pingFail) Ping(_ context.Context) error { return errors.New("unavailable") }

func TestHealthLive(t *testing.T) {
	h := NewHealth(health.NewReadiness(nil))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Live(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want ok", body["status"])
	}
}

func TestHealthReadyDatabaseOK(t *testing.T) {
	h := NewHealth(health.NewReadiness(pingOK{}))
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ready" || body.Checks["database"] != "ok" {
		t.Errorf("body = %+v", body)
	}
}

func TestHealthReadyDatabaseError(t *testing.T) {
	h := NewHealth(health.NewReadiness(pingFail{}))
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}

	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "not_ready" || body.Checks["database"] != "error" {
		t.Errorf("body = %+v", body)
	}
}
