package handlers

import (
	"net/http"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// Settings serves GET /api/v1/settings.
type Settings struct {
	view appsettings.PublicView
}

// SettingsDeps configures the settings handler.
type SettingsDeps struct {
	View appsettings.PublicView
}

// NewSettings creates a read-only settings handler.
func NewSettings(deps SettingsDeps) *Settings {
	return &Settings{view: deps.View}
}

func (s *Settings) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, s.view.Response())
}
