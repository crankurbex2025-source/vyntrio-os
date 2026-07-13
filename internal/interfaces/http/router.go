package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// SessionAuth wires session resolution and RBAC for protected routes.
type SessionAuth struct {
	Resolver   middleware.SessionResolver
	Authorizer ports.Authorizer
}

// NewRouter builds the HTTP router with middleware and routes.
func NewRouter(
	cfg config.Config,
	logger *slog.Logger,
	readiness *health.Readiness,
	bootstrap *handlers.Bootstrap,
	login *handlers.Login,
	logout *handlers.Logout,
	settings *handlers.Settings,
	sessionAuth *SessionAuth,
) http.Handler {
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
		if bootstrap != nil {
			r.Post("/identity/bootstrap", bootstrap.ServeHTTP)
		}
		if login != nil {
			r.Post("/identity/login", login.ServeHTTP)
		}
		if logout != nil && sessionAuth != nil && sessionAuth.Resolver != nil {
			r.With(
				middleware.OptionalAuthentication(sessionAuth.Resolver),
				middleware.RequireAuthentication,
				middleware.RequireCSRF,
			).Post("/identity/logout", logout.ServeHTTP)
		}
		if settings != nil && sessionAuth != nil && sessionAuth.Resolver != nil && sessionAuth.Authorizer != nil {
			r.With(
				middleware.OptionalAuthentication(sessionAuth.Resolver),
				middleware.RequireAuthentication,
				middleware.RequirePermission(sessionAuth.Authorizer, domainidentity.PermissionSettingsAdminRead),
			).Get("/settings", settings.ServeHTTP)
		}
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
