package middleware

import (
	"context"
	"errors"
	"net/http"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/ports"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/auth"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
)

// SessionResolver resolves raw session cookie values into authenticated subjects.
type SessionResolver interface {
	Resolve(ctx context.Context, rawSessionToken string) (appidentity.ResolvedSession, bool, error)
}

// OptionalAuthentication resolves a session when present and stores a principal in context.
func OptionalAuthentication(resolver SessionResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawSessionToken := ""
			if c, err := r.Cookie(cookie.SessionCookieName); err == nil {
				rawSessionToken = c.Value
			}

			session, ok, err := resolver.Resolve(r.Context(), rawSessionToken)
			if err != nil {
				writeAuthResolverError(w, r, err)
				return
			}

			ctx := r.Context()
			if ok {
				ctx = auth.WithPrincipal(ctx, auth.Principal{UserID: session.UserID, Role: session.Role})
				ctx = auth.WithSessionCSRF(ctx, auth.NewSessionCSRF(session.CSRFTokenHash))
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthentication rejects requests without a valid authenticated principal.
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.PrincipalFromContext(r.Context()); !ok {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required",
				GetRequestID(r.Context()))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission rejects requests that lack the requested permission.
func RequirePermission(authorizer ports.Authorizer, perm identity.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := GetRequestID(r.Context())
			principal, ok := auth.PrincipalFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", requestID)
				return
			}

			if err := authorizer.Authorize(principal.DomainPrincipal(), perm); err != nil {
				if errors.Is(err, identity.ErrUnauthorized) {
					response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", requestID)
					return
				}
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Permission denied", requestID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthResolverError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := GetRequestID(r.Context())
	if errors.Is(err, context.Canceled) {
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
		return
	}
	response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
}

var _ SessionResolver = (*appidentity.SessionResolver)(nil)
