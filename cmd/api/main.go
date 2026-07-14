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
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	appoverview "github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/ui"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/hostmetrics"
)

func main() {
	configPath, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger := newLogger(cfg)

	ctx := context.Background()
	store, err := sqlite.Open(ctx, cfg.StateDir)
	if err != nil {
		logger.Error("database startup failed", "error", err, "state_dir", cfg.StateDir)
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

	hasher, err := appidentity.NewPasswordHasher(appidentity.DefaultArgon2idConfig)
	if err != nil {
		_ = store.Close()
		logger.Error("password hasher init failed", "error", err)
		os.Exit(1)
	}
	userRepo := sqlite.NewUserRepository(store.DB())
	bootstrapRepo := sqlite.NewBootstrapRepository(store.DB())
	bootstrapService := appidentity.NewBootstrapService(hasher, bootstrapRepo, userRepo)
	bootstrapHandler := handlers.NewBootstrap(handlers.BootstrapDeps{Service: bootstrapService})

	sessionTokens, err := appidentity.NewSessionTokenService(appidentity.DefaultSessionTokenConfig)
	if err != nil {
		_ = store.Close()
		logger.Error("session token service init failed", "error", err)
		os.Exit(1)
	}
	loginRepo := sqlite.NewLoginRepository(store.DB())
	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())
	loginService := appidentity.NewLoginService(userRepo, hasher, sessionTokens, loginRepo, auditRepo)
	logoutRepo := sqlite.NewLogoutRepository(store.DB())
	logoutService := appidentity.NewLogoutService(logoutRepo)
	cookiePolicy := cookie.NewPolicy(cfg.CookieSecure)
	loginHandler := handlers.NewLogin(handlers.LoginDeps{
		Service:      loginService,
		CookiePolicy: cookiePolicy,
	})
	logoutHandler := handlers.NewLogout(handlers.LogoutDeps{
		Service:      logoutService,
		CookiePolicy: cookiePolicy,
	})

	sessionAuthRepo := sqlite.NewSessionAuthRepository(store.DB())
	sessionResolver := appidentity.NewSessionResolver(sessionAuthRepo)
	authorizer := ports.NewRBACAuthorizer()
	settingsLoader := appsettings.NewPublicSettingsLoader(settingsRepo, cfg.Version, cfg.Env)
	overviewLoader := appoverview.NewLoader(
		settingsRepo,
		readiness,
		hostmetrics.NewCollector(cfg.StateDir, hostmetrics.CollectorDeps{}),
		cfg.Version,
		cfg.BuildCommit,
		cfg.Env,
	)
	overviewHandler := handlers.NewOverview(handlers.OverviewDeps{Loader: overviewLoader})
	settingsHandler := handlers.NewSettings(handlers.SettingsDeps{Loader: settingsLoader})
	instanceDisplayNameRepo := sqlite.NewInstanceDisplayNameRepository(store.DB())
	updateInstanceService := appsettings.NewUpdateInstanceDisplayNameService(instanceDisplayNameRepo)
	updateInstanceHandler := handlers.NewUpdateInstanceSettings(handlers.UpdateInstanceSettingsDeps{
		Service: updateInstanceService,
	})

	uiHandler, err := ui.NewHandler()
	if err != nil {
		_ = store.Close()
		logger.Error("embedded ui init failed", "error", err)
		os.Exit(1)
	}

	srv := httpapi.NewServer(
		cfg,
		logger,
		readiness,
		bootstrapHandler,
		loginHandler,
		logoutHandler,
		overviewHandler,
		settingsHandler,
		updateInstanceHandler,
		&httpapi.SessionAuth{
			Resolver:   sessionResolver,
			Authorizer: authorizer,
		},
		httpapi.WithUI(uiHandler),
	)

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
