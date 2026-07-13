package httpapi_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func TestShutdownWithoutActiveListener(t *testing.T) {
	cfg := config.Config{
		APIHost:         "127.0.0.1",
		APIPort:         8080,
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		IdleTimeout:     time.Second,
		ShutdownTimeout: time.Second,
		Version:         "test",
		BuildCommit:     "test",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httpapi.NewServer(cfg, logger, health.NewReadiness(nil), nil, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
}
