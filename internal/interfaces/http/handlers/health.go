package handlers

import (
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// Health handles liveness and readiness probes.
type Health struct {
	readiness *health.Readiness
}

// NewHealth creates health handlers with the given readiness evaluator.
func NewHealth(readiness *health.Readiness) *Health {
	return &Health{readiness: readiness}
}

func (h *Health) Live(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	result := h.readiness.Check(r.Context())

	checks := map[string]string{"process": "ok"}
	status := "ready"
	code := http.StatusOK

	if result.DatabaseOK {
		checks["database"] = "ok"
	} else {
		checks["database"] = "error"
		status = "not_ready"
		code = http.StatusServiceUnavailable
	}

	response.JSON(w, code, map[string]any{
		"status": status,
		"checks": checks,
	})
}
