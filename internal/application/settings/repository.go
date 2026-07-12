// Package settings defines application ports for system settings persistence.
package settings

import (
	"context"
	"errors"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

// ErrNotFound indicates the requested setting does not exist.
var ErrNotFound = errors.New("setting not found")

// Repository persists namespaced settings (Slice 2b.1: system namespace only).
type Repository interface {
	Get(ctx context.Context, namespace, key string) (setting.Setting, error)
	Set(ctx context.Context, s setting.Setting) error
	ListByNamespace(ctx context.Context, namespace string) ([]setting.Setting, error)
}
