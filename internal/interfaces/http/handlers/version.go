package handlers

import (
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// VersionInfo is returned by GET /api/v1/version.
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// Version serves build metadata.
type Version struct {
	info VersionInfo
}

func NewVersion(version, commit string) *Version {
	return &Version{info: VersionInfo{Version: version, Commit: commit}}
}

func (v *Version) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, v.info)
}
