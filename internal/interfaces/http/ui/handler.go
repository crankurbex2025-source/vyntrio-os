package ui

import (
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	htmlCacheControl  = "no-cache"
	assetCacheControl = "public, max-age=31536000, immutable"

	contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; " +
		"img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; " +
		"form-action 'self'; frame-ancestors 'none'"

	assetsURLPrefix = "/assets/"
	distIndexPath   = "dist/index.html"
	distAssetsDir   = "dist/assets"
)

// Handler serves the embedded production frontend: index.html for UI
// entry/fallback paths and immutable static files beneath /assets/.
type Handler struct {
	index []byte
}

// NewHandler validates the embedded production assets and returns a UI handler.
func NewHandler() (*Handler, error) {
	index, err := distFS.ReadFile(distIndexPath)
	if err != nil {
		return nil, fmt.Errorf("embedded ui: read %s: %w", distIndexPath, err)
	}
	if len(index) == 0 {
		return nil, errors.New("embedded ui: dist/index.html is empty")
	}

	entries, err := fs.ReadDir(distFS, distAssetsDir)
	if err != nil {
		return nil, fmt.Errorf("embedded ui: read %s: %w", distAssetsDir, err)
	}
	if len(entries) == 0 {
		return nil, errors.New("embedded ui: dist/assets contains no files")
	}

	return &Handler{index: index}, nil
}

// ServeIndex serves the embedded index.html for GET/HEAD UI entry and
// SPA-fallback requests with UI cache and security headers.
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeStaticNotFound(w)
		return
	}

	header := w.Header()
	header.Set("Cache-Control", htmlCacheControl)
	header.Set("Content-Type", "text/html; charset=utf-8")
	header.Set("Content-Security-Policy", contentSecurityPolicy)
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("Content-Length", strconv.Itoa(len(h.index)))
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodHead {
		_, _ = w.Write(h.index)
	}
}

// ServeAsset serves one embedded static file beneath /assets/ for GET/HEAD
// requests with immutable cache headers. Anything else is a static 404.
func (h *Handler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeStaticNotFound(w)
		return
	}

	name, ok := assetName(r.URL.Path)
	if !ok {
		writeStaticNotFound(w)
		return
	}

	data, err := distFS.ReadFile(distAssetsDir + "/" + name)
	if err != nil {
		writeStaticNotFound(w)
		return
	}

	contentType := assetContentType(name)

	header := w.Header()
	header.Set("Content-Type", contentType)
	header.Set("Cache-Control", assetCacheControl)
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

// assetContentType returns a deterministic Content-Type for known Vite
// output extensions, independent of host mime tables.
func assetContentType(name string) string {
	switch path.Ext(name) {
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".map", ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	case ".woff2":
		return "font/woff2"
	default:
		if t := mime.TypeByExtension(path.Ext(name)); t != "" {
			return t
		}
		return "application/octet-stream"
	}
}

// assetName extracts and validates the embedded file name from a decoded
// /assets/ URL path. It rejects traversal, backslashes, dot segments,
// dot-files, empty segments, and directory requests.
func assetName(urlPath string) (string, bool) {
	if !strings.HasPrefix(urlPath, assetsURLPrefix) {
		return "", false
	}
	name := urlPath[len(assetsURLPrefix):]
	if name == "" || strings.HasSuffix(name, "/") || strings.Contains(name, "\\") {
		return "", false
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || strings.HasPrefix(segment, ".") {
			return "", false
		}
	}
	return name, true
}

func writeStaticNotFound(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Content-Type", "text/plain; charset=utf-8")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("404 page not found\n"))
}
