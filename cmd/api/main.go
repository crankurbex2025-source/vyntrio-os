// Package main is the HTTP API service entrypoint for Vyntrio OS.
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	logger.Info("database ready", "path", store.Path())

	settingsRepo := sqlite.NewSettingsRepository(store.DB())
	sysSettings, err := appsettings.NewReader(settingsRepo).LoadSystemSettings(ctx)
	if err != nil {
		_ = store.Close()
		logger.Error("system settings load failed", "error", err)
		os.Exit(1)
	}
	logger = appsettings.LoggerWithSystemSettings(logger, sysSettings, strings.EqualFold(cfg.LogLevel, "debug"))

	snapshot := appsettings.NewSnapshot(sysSettings)
	if err := appsettings.VerifyPersisted(ctx, settingsRepo, snapshot); err != nil {
		_ = store.Close()
		logger.Error("system settings persistence verification failed", "error", err)
		os.Exit(1)
	}

	readiness := health.NewReadiness(store)
	srv := httpapi.NewServer(cfg, logger, readiness)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		_ = store.Close()
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = store.Close()
		logger.Error("http shutdown failed", "error", err)
		os.Exit(1)
	}

	if err := store.Close(); err != nil {
		logger.Error("database close failed", "error", err)
		os.Exit(1)
	}

	logger.Info("shutdown complete")
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
