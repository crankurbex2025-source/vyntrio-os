package handlers

import (
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// Health handles liveness and readiness probes.
type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) Live(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) Ready(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"status": "ready",
		"checks": map[string]string{
			"process": "ok",
		},
	})
}
