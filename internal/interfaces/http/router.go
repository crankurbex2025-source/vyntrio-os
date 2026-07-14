package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/ui"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// SessionAuth wires session resolution and RBAC for protected routes.
type SessionAuth struct {
	Resolver   middleware.SessionResolver
	Authorizer ports.Authorizer
}

// RouterOption configures optional router behavior.
type RouterOption func(*routerOptions)

type routerOptions struct {
	ui *ui.Handler
}

// WithUI enables embedded production frontend serving: /assets/* static
// files plus index.html entry/SPA fallback for non-reserved GET/HEAD paths.
func WithUI(handler *ui.Handler) RouterOption {
	return func(o *routerOptions) {
		o.ui = handler
	}
}

// NewRouter builds the HTTP router with middleware and routes.
func NewRouter(
	cfg config.Config,
	logger *slog.Logger,
	readiness *health.Readiness,
	bootstrap *handlers.Bootstrap,
	login *handlers.Login,
	logout *handlers.Logout,
	overview *handlers.Overview,
	settings *handlers.Settings,
	updateInstance *handlers.UpdateInstanceSettings,
	sessionAuth *SessionAuth,
	opts ...RouterOption,
) http.Handler {
	var options routerOptions
	for _, opt := range opts {
		opt(&options)
	}

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
		if overview != nil && sessionAuth != nil && sessionAuth.Resolver != nil && sessionAuth.Authorizer != nil {
			r.With(
				middleware.OptionalAuthentication(sessionAuth.Resolver),
				middleware.RequireAuthentication,
				middleware.RequirePermission(sessionAuth.Authorizer, domainidentity.PermissionSystemHealth),
			).Get("/overview", overview.ServeHTTP)
		}
		if settings != nil && sessionAuth != nil && sessionAuth.Resolver != nil && sessionAuth.Authorizer != nil {
			r.With(
				middleware.OptionalAuthentication(sessionAuth.Resolver),
				middleware.RequireAuthentication,
				middleware.RequirePermission(sessionAuth.Authorizer, domainidentity.PermissionSettingsAdminRead),
			).Get("/settings", settings.ServeHTTP)
		}
		if updateInstance != nil && sessionAuth != nil && sessionAuth.Resolver != nil && sessionAuth.Authorizer != nil {
			r.With(
				middleware.OptionalAuthentication(sessionAuth.Resolver),
				middleware.RequireAuthentication,
				middleware.RequirePermission(sessionAuth.Authorizer, domainidentity.PermissionSettingsAdminWrite),
				middleware.RequireCSRF,
			).Patch("/settings/instance", updateInstance.ServeHTTP)
		}
	})

	if options.ui != nil {
		r.Get("/assets/*", options.ui.ServeAsset)
		r.Head("/assets/*", options.ui.ServeAsset)
		r.NotFound(uiFallbackHandler(options.ui))
	} else {
		r.NotFound(notFoundHandler)
	}
	r.MethodNotAllowed(methodNotAllowedHandler)

	return r
}

// uiFallbackHandler serves the embedded index.html for non-reserved GET/HEAD
// paths and preserves the canonical JSON 404 for everything else, including
// every /api path, operational probes, and static asset paths.
func uiFallbackHandler(uiHandler *ui.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isUIFallbackRequest(r) {
			notFoundHandler(w, r)
			return
		}
		uiHandler.ServeIndex(w, r)
	}
}

func isUIFallbackRequest(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	p := r.URL.Path
	if p == "/api" || strings.HasPrefix(p, "/api/") {
		return false
	}
	if p == "/assets" || strings.HasPrefix(p, "/assets/") {
		return false
	}
	if p == "/healthz" || p == "/readyz" {
		return false
	}
	return true
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found",
		middleware.GetRequestID(r.Context()))
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed",
		middleware.GetRequestID(r.Context()))
}
