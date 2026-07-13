package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/auth"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/request"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/google/uuid"
)

const updateInstanceMaxBodyBytes = 4 * 1024

type updateInstanceRequest struct {
	DisplayName string `json:"display_name"`
}

type updateInstanceResponse struct {
	DisplayName string `json:"display_name"`
}

// UpdateInstanceSettings handles PATCH /api/v1/settings/instance.
type UpdateInstanceSettings struct {
	service    *appsettings.UpdateInstanceDisplayNameService
	newAuditID func() string
}

// UpdateInstanceSettingsDeps configures the instance settings update handler.
type UpdateInstanceSettingsDeps struct {
	Service *appsettings.UpdateInstanceDisplayNameService
}

// NewUpdateInstanceSettings creates an instance settings update handler.
func NewUpdateInstanceSettings(deps UpdateInstanceSettingsDeps) *UpdateInstanceSettings {
	return &UpdateInstanceSettings{
		service:    deps.Service,
		newAuditID: func() string { return uuid.NewString() },
	}
}

func (h *UpdateInstanceSettings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	if err := r.Context().Err(); err != nil {
		response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
		return
	}

	var payload updateInstanceRequest
	if err := request.DecodeStrictJSON(r, updateInstanceMaxBodyBytes, &payload); err != nil {
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

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", requestID)
		return
	}

	result, err := h.service.Update(r.Context(), principal.UserID, payload.DisplayName, h.newAuditID())
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			response.Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out", requestID)
		case appsettings.IsInvalidInstanceDisplayNameError(err):
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request", requestID)
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		}
		return
	}

	response.JSON(w, http.StatusOK, updateInstanceResponse{DisplayName: result.DisplayName})
}
