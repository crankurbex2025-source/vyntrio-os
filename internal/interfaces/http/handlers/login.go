package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/request"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/google/uuid"
)

const loginMaxBodyBytes = 8 * 1024

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles username/password authentication requests.
type Login struct {
	service      *appidentity.LoginService
	cookiePolicy cookie.Policy
	now          func() time.Time
	newSessionID func() string
	newAuditID   func() string
}

// LoginDeps configures the login handler.
type LoginDeps struct {
	Service      *appidentity.LoginService
	CookiePolicy cookie.Policy
}

// NewLogin creates a login handler.
func NewLogin(deps LoginDeps) *Login {
	return &Login{
		service:      deps.Service,
		cookiePolicy: deps.CookiePolicy,
		now:          func() time.Time { return time.Now().UTC() },
		newSessionID: func() string { return uuid.NewString() },
		newAuditID:   func() string { return uuid.NewString() },
	}
}

func (h *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	if err := r.Context().Err(); err != nil {
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
		return
	}

	var payload loginRequest
	if err := request.DecodeStrictJSON(r, loginMaxBodyBytes, &payload); err != nil {
		switch {
		case errors.Is(err, request.ErrInvalidContentType),
			errors.Is(err, request.ErrInvalidJSON),
			errors.Is(err, request.ErrBodyTooLarge),
			errors.Is(err, io.EOF):
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request", requestID)
		default:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request", requestID)
		}
		return
	}

	result, err := h.service.Login(
		r.Context(),
		payload.Username,
		payload.Password,
		h.newSessionID(),
		h.newAuditID(),
	)
	if err != nil {
		h.writeLoginError(w, requestID, err)
		return
	}

	h.cookiePolicy.SetSessionCookie(w, result.Material, h.now())
	response.NoContent(w)
}

func (h *Login) writeLoginError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
	case appidentity.IsLoginClientInputError(err):
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request", requestID)
	case appidentity.IsLoginAuthenticationError(err):
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", requestID)
	default:
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
	}
}
