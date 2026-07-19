package handlers

import (
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installmediapublic"
)

// PublicInstallMedia serves GET /api/v1/public/install-media.
type PublicInstallMedia struct {
	reader installmediapublic.Reader
}

// PublicInstallMediaDeps configures the public install-media handler.
type PublicInstallMediaDeps struct {
	Reader installmediapublic.Reader
}

// NewPublicInstallMedia creates the public install-media metadata handler.
func NewPublicInstallMedia(deps PublicInstallMediaDeps) *PublicInstallMedia {
	return &PublicInstallMedia{reader: deps.Reader}
}

func (h *PublicInstallMedia) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	w.Header().Set("Cache-Control", "no-store")

	payload, err := h.reader.Read()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	response.JSON(w, http.StatusOK, payload)
}
