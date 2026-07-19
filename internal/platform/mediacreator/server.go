package mediacreator

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/writemedia"
)

const defaultListen = "127.0.0.1:17823"

// Options configures the local media-creator GUI server.
type Options struct {
	Version     string
	Listen      string
	OpenBrowser bool
	ImageHint   string
}

// Run starts the loopback GUI server and blocks until the context is cancelled
// or the process receives a shutdown via /api/shutdown.
func Run(ctx context.Context, opts Options) error {
	logPath := StartupLogPath()
	if strings.TrimSpace(opts.Listen) == "" {
		opts.Listen = defaultListen
	}
	if strings.TrimSpace(opts.Version) == "" {
		opts.Version = "0.2.0-dev"
	}

	AppendLog(logPath, fmt.Sprintf("start version=%s listen=%s goos=%s goarch=%s", opts.Version, opts.Listen, runtime.GOOS, runtime.GOARCH))

	srv := &server{
		version:   opts.Version,
		imageHint: strings.TrimSpace(opts.ImageHint),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/assets/app.css", srv.handleCSS)
	mux.HandleFunc("/assets/app.js", srv.handleJS)
	mux.HandleFunc("/api/status", srv.handleStatus)
	mux.HandleFunc("/api/devices", srv.handleDevices)
	mux.HandleFunc("/api/image", srv.handleImage)
	mux.HandleFunc("/api/suggest", srv.handleSuggest)
	mux.HandleFunc("/api/write", srv.handleWrite)
	mux.HandleFunc("/api/verify", srv.handleVerify)
	mux.HandleFunc("/api/shutdown", srv.handleShutdown)

	listener, err := net.Listen("tcp", opts.Listen)
	if err != nil {
		// Fall back to an ephemeral port if the preferred port is busy.
		if opts.Listen == defaultListen {
			AppendLog(logPath, fmt.Sprintf("listen %s failed (%v); falling back to 127.0.0.1:0", opts.Listen, err))
			listener, err = net.Listen("tcp", "127.0.0.1:0")
		}
		if err != nil {
			AppendLog(logPath, "listen failed: "+err.Error())
			userFacingStartupError(err, logPath)
			return fmt.Errorf("listen: %w", err)
		}
	}
	addr := listener.Addr().String()
	url := "http://" + addr + "/"
	AppendLog(logPath, "listening "+url)

	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	srv.httpServer = httpServer

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	fmt.Fprintf(os.Stderr, "vyntrio-media-creator: local GUI at %s\n", url)
	fmt.Fprintf(os.Stderr, "vyntrio-media-creator: log %s\n", logPath)
	fmt.Fprintln(os.Stderr, "vyntrio-media-creator: loopback only — destructive USB write requires elevation")

	browserErr := error(nil)
	if opts.OpenBrowser {
		browserErr = openBrowser(url)
		if browserErr != nil {
			AppendLog(logPath, "open browser failed: "+browserErr.Error())
		} else {
			AppendLog(logPath, "open browser requested")
		}
	}
	notifyLaunch(url, logPath)

	select {
	case <-ctx.Done():
		AppendLog(logPath, "shutdown: context done")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if err == nil || err == http.ErrServerClosed {
			AppendLog(logPath, "server closed")
			return nil
		}
		AppendLog(logPath, "server error: "+err.Error())
		return err
	case <-srv.done():
		AppendLog(logPath, "shutdown: api")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		return nil
	}
}

type server struct {
	version     string
	imageHint   string
	httpServer  *http.Server
	mu          sync.Mutex
	writing     bool
	shutdownOnce sync.Once
	shutdownCh  chan struct{}
}

func (s *server) done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shutdownCh == nil {
		s.shutdownCh = make(chan struct{})
	}
	return s.shutdownCh
}

func (s *server) requestShutdown() {
	s.shutdownOnce.Do(func() {
		s.mu.Lock()
		if s.shutdownCh == nil {
			s.shutdownCh = make(chan struct{})
		}
		ch := s.shutdownCh
		s.mu.Unlock()
		close(ch)
	})
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *server) handleCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(appCSS))
}

func (s *server) handleJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(appJS))
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":            "Vyntrio Media Creator",
		"version":         s.version,
		"kind":            "local_web_gui",
		"native_gui":      false,
		"support_status":  "engineering_media_early_access",
		"image_hint":      s.imageHint,
		"artifact_name":   "vyntrio-install-media.img",
		"platform":        runtime.GOOS,
		"arch":            runtime.GOARCH,
		"requires_elevation": true,
		"boot_instruction": "Boot the USB in UEFI or BIOS/legacy mode (dual-mode GPT hybrid).",
	})
}

func (s *server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	devices, err := writemedia.ListDevices()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": devices})
}

func (s *server) handleImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	img, err := writemedia.LoadImage(strings.TrimSpace(body.Path))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":             img.Name,
		"path":             img.Path,
		"size_bytes":       img.SizeBytes,
		"sha256":           img.SHA256,
		"expected_sha256":  img.ExpectedSHA256,
		"manifest_path":    img.ManifestPath,
	})
}

func (s *server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	candidates := suggestImagePaths()
	writeJSON(w, http.StatusOK, map[string]any{"candidates": candidates})
}

func (s *server) handleWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var body struct {
		ImagePath string `json:"image_path"`
		Device    string `json:"device"`
		Confirm   bool   `json:"confirm"`
		DryRun    bool   `json:"dry_run"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if !body.Confirm && !body.DryRun {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "confirm must be true for destructive write"})
		return
	}

	s.mu.Lock()
	if s.writing {
		s.mu.Unlock()
		writeJSON(w, http.StatusConflict, map[string]any{"error": "write already in progress"})
		return
	}
	s.writing = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.writing = false
		s.mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "streaming unsupported"})
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	encode := func(payload map[string]any) {
		_ = json.NewEncoder(w).Encode(payload)
		flusher.Flush()
	}
	encode(map[string]any{"type": "started", "dry_run": body.DryRun})

	result, err := writemedia.WriteImage(body.ImagePath, body.Device, writemedia.WriteOptions{
		AssumeYes: body.Confirm,
		DryRun:    body.DryRun,
		OnProgress: func(done, total uint64) {
			encode(map[string]any{
				"type":        "progress",
				"bytes_done":  done,
				"total_bytes": total,
			})
		},
	})
	if err != nil {
		encode(map[string]any{"type": "error", "error": err.Error()})
		return
	}
	encode(map[string]any{
		"type":          "complete",
		"device_path":   result.DevicePath,
		"image_path":    result.ImagePath,
		"bytes_written": result.BytesWritten,
		"verified":      result.Verified,
		"boot_instruction": "Remove the USB safely, insert it into the target machine, and boot in UEFI or BIOS/legacy mode.",
	})
}

func (s *server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var body struct {
		ImagePath string `json:"image_path"`
		Device    string `json:"device"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if err := writemedia.VerifyDevice(body.ImagePath, body.Device); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "verified": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"verified": true})
}

func (s *server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	go func() {
		time.Sleep(200 * time.Millisecond)
		s.requestShutdown()
	}()
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func suggestImagePaths() []string {
	var roots []string
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, cwd)
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots,
			filepath.Join(home, "Downloads"),
			filepath.Join(home, "Download"),
			filepath.Join(home, "Desktop"),
		)
	}
	seen := map[string]struct{}{}
	var out []string
	for _, root := range roots {
		matches, _ := filepath.Glob(filepath.Join(root, "vyntrio-install-media*.img"))
		for _, match := range matches {
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			out = append(out, match)
		}
	}
	return out
}
