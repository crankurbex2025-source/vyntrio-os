// Package main is the HTTP API service entrypoint for Vyntrio OS.
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger := newLogger(cfg)

	ctx := context.Background()
	store, err := sqlite.Open(ctx, cfg.DataDir)
	if err != nil {
		logger.Error("database startup failed", "error", err, "data_dir", cfg.DataDir)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Error("database close failed", "error", err)
		}
	}()

	logger.Info("database ready", "path", store.Path())

	settingsRepo := sqlite.NewSettingsRepository(store.DB())
	sysSettings, err := appsettings.NewReader(settingsRepo).LoadSystemSettings(ctx)
	if err != nil {
		logger.Error("system settings load failed", "error", err)
		os.Exit(1)
	}
	logger = appsettings.LoggerWithSystemSettings(logger, sysSettings, strings.EqualFold(cfg.LogLevel, "debug"))

	readiness := health.NewReadiness(store)
	srv := httpapi.NewServer(cfg, logger, readiness)

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.Env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
