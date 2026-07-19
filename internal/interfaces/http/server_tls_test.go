package httpapi_test

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/tlsutil"
)

func TestListenAndServeTLSHealthz(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "var", "lib", "vyntrio")
	if err := os.MkdirAll(stateDir, 0o750); err != nil {
		t.Fatal(err)
	}

	certPath, keyPath, err := tlsutil.WriteSelfSignedFiles(filepath.Join(stateDir, "tls"), "127.0.0.1")
	if err != nil {
		t.Fatalf("WriteSelfSignedFiles() error: %v", err)
	}

	configPath := filepath.Join(t.TempDir(), "config.toml")
	body := strings.TrimSpace(`
bind_address = "127.0.0.1"
listen_port = 38443
state_dir = "`+stateDir+`"
log_level = "info"
cookie_secure = true
tls_cert_file = "`+certPath+`"
tls_key_file = "`+keyPath+`"
`) + "\n"
	if err := os.WriteFile(configPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadWithOptions(configPath, config.LoadOptions{AllowedStateDir: stateDir})
	if err != nil {
		t.Fatalf("LoadWithOptions() error: %v", err)
	}
	cfg.ReadTimeout = time.Second
	cfg.WriteTimeout = time.Second
	cfg.IdleTimeout = time.Second
	cfg.ShutdownTimeout = time.Second
	cfg.Version = "test"
	cfg.BuildCommit = "test"

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httpapi.NewServer(cfg, logger, health.NewReadiness(nil), nil, nil, nil, nil, nil, nil, nil, nil)

	go func() {
		_ = srv.ListenAndServe()
	}()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test-only self-signed probe
		},
		Timeout: time.Second,
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get("https://" + srv.Addr() + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					t.Fatalf("Shutdown() error: %v", err)
				}
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("HTTPS /healthz did not return 200 on %s", srv.Addr())
}
