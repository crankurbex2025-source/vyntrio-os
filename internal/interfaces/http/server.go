package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

// Server wraps the HTTP server for cmd/api.
type Server struct {
	cfg    config.Config
	logger *slog.Logger
	http   *http.Server
}

// NewServer creates an configured HTTP server (not started).
func NewServer(
	cfg config.Config,
	logger *slog.Logger,
	readiness *health.Readiness,
	bootstrap *handlers.Bootstrap,
	login *handlers.Login,
	logout *handlers.Logout,
	overview *handlers.Overview,
	settings *handlers.Settings,
	updateInstance *handlers.UpdateInstanceSettings,
	storage *handlers.Storage,
	sessionAuth *SessionAuth,
	opts ...RouterOption,
) *Server {
	handler := NewRouter(cfg, logger, readiness, bootstrap, login, logout, overview, settings, updateInstance, storage, sessionAuth, opts...)
	return &Server{
		cfg:    cfg,
		logger: logger,
		http: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// ListenAndServe starts the HTTP or HTTPS server.
func (s *Server) ListenAndServe() error {
	if s.cfg.TLSEnabled() {
		s.logger.Info("api server listening", "addr", s.cfg.Addr(), "env", s.cfg.Env, "tls", true)
		return s.http.ListenAndServeTLS(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
	}
	s.logger.Info("api server listening", "addr", s.cfg.Addr(), "env", s.cfg.Env, "tls", false)
	return s.http.ListenAndServe()
}

// Handler returns the root handler (for tests).
func (s *Server) Handler() http.Handler {
	return s.http.Handler
}

// Addr returns the listen address.
func (s *Server) Addr() string {
	return s.cfg.Addr()
}

// Shutdown gracefully stops the HTTP server and waits for in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
