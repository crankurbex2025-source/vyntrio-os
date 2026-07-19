package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installmediapublic"
	"github.com/go-chi/chi/v5"
)

var allowedReleaseFiles = map[string]struct{}{
	installmediapublic.ArtifactName:           {},
	installmediapublic.ArtifactNameLegacyBIOS: {},
	installmediapublic.ManifestName:           {},
}

func allowedWriterFiles() map[string]struct{} {
	files := make(map[string]struct{}, len(installmediapublic.WriterArtifactNames)*2)
	for _, artifact := range installmediapublic.WriterArtifactNames {
		files[artifact.Name] = struct{}{}
		files[artifact.Name+".sha256"] = struct{}{}
	}
	return files
}

// ReleaseFiles serves allowlisted files from the release staging directory.
type ReleaseFiles struct {
	stagingDir string
}

// NewReleaseFiles creates a release file handler.
func NewReleaseFiles(stagingDir string) *ReleaseFiles {
	return &ReleaseFiles{stagingDir: stagingDir}
}

func (h *ReleaseFiles) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	if h.stagingDir == "" {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", requestID)
		return
	}

	filename := chi.URLParam(r, "filename")
	if _, ok := allowedReleaseFiles[filename]; !ok {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", requestID)
		return
	}

	path := filepath.Join(h.stagingDir, filename)
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, path)
}

// ServeWriter serves allowlisted writer binaries from the staging writer subdirectory.
func (h *ReleaseFiles) ServeWriter(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	if h.stagingDir == "" {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", requestID)
		return
	}

	filename := chi.URLParam(r, "filename")
	if _, ok := allowedWriterFiles()[filename]; !ok {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", requestID)
		return
	}

	path := filepath.Join(h.stagingDir, installmediapublic.WriterStagingSubdir, filename)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	http.ServeFile(w, r, path)
}
