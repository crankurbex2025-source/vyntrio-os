package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/peer"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/request"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/google/uuid"
)

const bootstrapMaxBodyBytes = 8 * 1024

type bootstrapRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type bootstrapResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// Bootstrap handles first-owner bootstrap requests.
type Bootstrap struct {
	service    *appidentity.BootstrapService
	newUserID  func() string
	newAuditID func() string
}

// BootstrapDeps configures the bootstrap handler.
type BootstrapDeps struct {
	Service *appidentity.BootstrapService
}

// NewBootstrap creates a bootstrap handler.
func NewBootstrap(deps BootstrapDeps) *Bootstrap {
	return &Bootstrap{
		service:    deps.Service,
		newUserID:  func() string { return uuid.NewString() },
		newAuditID: func() string { return uuid.NewString() },
	}
}

func (h *Bootstrap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	if !peer.IsLoopback(r.RemoteAddr) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", requestID)
		return
	}

	if err := r.Context().Err(); err != nil {
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
		return
	}

	var payload bootstrapRequest
	if err := request.DecodeStrictJSON(r, bootstrapMaxBodyBytes, &payload); err != nil {
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

	result, err := h.service.CreateFirstOwner(
		r.Context(),
		payload.Username,
		payload.Password,
		h.newUserID(),
		h.newAuditID(),
	)
	if err != nil {
		h.writeBootstrapError(w, requestID, err)
		return
	}

	response.JSON(w, http.StatusCreated, bootstrapResponse{
		UserID:   result.UserID,
		Username: result.Username,
		Role:     result.Role,
		Status:   result.Status,
	})
}

func (h *Bootstrap) writeBootstrapError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
	case appidentity.IsBootstrapClientInputError(err):
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request", requestID)
	case appidentity.IsBootstrapUnavailableError(err):
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", requestID)
	default:
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
	}
}
