package handlers

import (
	"net/http"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// Settings serves GET /api/v1/settings.
type Settings struct {
	loader appsettings.PublicSettingsLoader
}

// SettingsDeps configures the settings handler.
type SettingsDeps struct {
	Loader appsettings.PublicSettingsLoader
}

// NewSettings creates a read-only settings handler.
func NewSettings(deps SettingsDeps) *Settings {
	return &Settings{loader: deps.Loader}
}

func (s *Settings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	payload, err := s.loader.Load(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	response.JSON(w, http.StatusOK, payload)
}
