package handlers

import (
	"context"
	"errors"
	"net/http"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/google/uuid"
)

// Logout handles session logout requests.
type Logout struct {
	service      *appidentity.LogoutService
	cookiePolicy cookie.Policy
	newAuditID   func() string
}

// LogoutDeps configures the logout handler.
type LogoutDeps struct {
	Service      *appidentity.LogoutService
	CookiePolicy cookie.Policy
}

// NewLogout creates a logout handler.
func NewLogout(deps LogoutDeps) *Logout {
	return &Logout{
		service:      deps.Service,
		cookiePolicy: deps.CookiePolicy,
		newAuditID:   func() string { return uuid.NewString() },
	}
}

func (h *Logout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	if err := r.Context().Err(); err != nil {
		h.cookiePolicy.ClearSessionCookie(w)
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
		return
	}

	rawSessionToken := ""
	if c, err := r.Cookie(cookie.SessionCookieName); err == nil {
		rawSessionToken = c.Value
	}

	_, err := h.service.Logout(r.Context(), rawSessionToken, h.newAuditID())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			h.cookiePolicy.ClearSessionCookie(w)
			response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
			return
		}
		h.cookiePolicy.ClearSessionCookie(w)
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	h.cookiePolicy.ClearSessionCookie(w)
	response.NoContent(w)
}
