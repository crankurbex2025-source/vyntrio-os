package settings

import (
	"context"
	"errors"
	"fmt"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

// ErrInvalidInstanceDisplayName indicates the display name failed validation.
var ErrInvalidInstanceDisplayName = errors.New("invalid instance display name")

// InstanceDisplayNameStore atomically updates the persisted instance display name with audit.
type InstanceDisplayNameStore interface {
	UpdateInstanceDisplayNameWithAudit(
		ctx context.Context,
		displayName string,
		actorUserID domainidentity.UserID,
		auditID string,
	) (changed bool, err error)
}

// UpdateInstanceDisplayNameResult describes the outcome of an instance display name update.
type UpdateInstanceDisplayNameResult struct {
	DisplayName string
	Changed     bool
}

// UpdateInstanceDisplayNameService updates the canonical instance display name.
type UpdateInstanceDisplayNameService struct {
	store InstanceDisplayNameStore
}

// NewUpdateInstanceDisplayNameService returns an instance display name update service.
func NewUpdateInstanceDisplayNameService(store InstanceDisplayNameStore) *UpdateInstanceDisplayNameService {
	return &UpdateInstanceDisplayNameService{store: store}
}

// Update validates and persists the instance display name when changed.
func (s *UpdateInstanceDisplayNameService) Update(
	ctx context.Context,
	actorUserID domainidentity.UserID,
	rawDisplayName, auditID string,
) (UpdateInstanceDisplayNameResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateInstanceDisplayNameResult{}, err
	}

	displayName, err := setting.ValidateInstanceDisplayName(rawDisplayName)
	if err != nil {
		return UpdateInstanceDisplayNameResult{}, fmt.Errorf("%w: %v", ErrInvalidInstanceDisplayName, err)
	}

	changed, err := s.store.UpdateInstanceDisplayNameWithAudit(ctx, displayName, actorUserID, auditID)
	if err != nil {
		return UpdateInstanceDisplayNameResult{}, err
	}

	return UpdateInstanceDisplayNameResult{
		DisplayName: displayName,
		Changed:     changed,
	}, nil
}

// IsInvalidInstanceDisplayNameError reports client input validation failures.
func IsInvalidInstanceDisplayNameError(err error) bool {
	return errors.Is(err, ErrInvalidInstanceDisplayName)
}
