package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds the HTTP router with middleware and routes.
func NewRouter(cfg config.Config, logger *slog.Logger, readiness *health.Readiness) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(chimiddleware.Timeout(cfg.ReadTimeout))

	healthHandler := handlers.NewHealth(readiness)
	r.Get("/healthz", healthHandler.Live)
	r.Get("/readyz", healthHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", handlers.NewVersion(cfg.Version, cfg.BuildCommit).ServeHTTP)
	})

	r.NotFound(notFoundHandler)
	r.MethodNotAllowed(methodNotAllowedHandler)

	return r
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found",
		middleware.GetRequestID(r.Context()))
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed",
		middleware.GetRequestID(r.Context()))
}
