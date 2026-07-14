package handlers

import (
	"net/http"

	appoverview "github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// Overview serves GET /api/v1/overview.
type Overview struct {
	loader appoverview.Loader
}

// OverviewDeps configures the overview handler.
type OverviewDeps struct {
	Loader appoverview.Loader
}

// NewOverview creates a read-only overview handler.
func NewOverview(deps OverviewDeps) *Overview {
	return &Overview{loader: deps.Loader}
}

func (h *Overview) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	w.Header().Set("Cache-Control", "no-store")

	payload, err := h.loader.Load(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	response.JSON(w, http.StatusOK, payload)
}
