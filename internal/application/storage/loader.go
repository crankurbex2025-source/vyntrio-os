package storage

import (
	"context"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

// InventorySource loads the block device inventory DTO.
type InventorySource interface {
	Load(ctx context.Context) (DisksResponse, error)
}

// InventoryLoader assembles the storage inventory API response.
type InventoryLoader struct {
	collector storageinventory.Collector
}

// NewInventoryLoader creates a loader for GET /api/v1/storage/disks.
func NewInventoryLoader(collector storageinventory.Collector) InventoryLoader {
	return InventoryLoader{collector: collector}
}

// Load returns the current read-only storage inventory DTO.
func (l InventoryLoader) Load(ctx context.Context) (DisksResponse, error) {
	_ = ctx
	inventory := l.collector.Collect(ctx)
	return mapInventory(inventory), nil
}

func mapInventory(inventory storageinventory.Inventory) DisksResponse {
	devices := make([]DiskDevice, 0, len(inventory.Devices))
	for _, device := range inventory.Devices {
		devices = append(devices, DiskDevice{
			ID:          device.ID,
			Status:      device.Status,
			SizeBytes:   device.SizeBytes,
			Rotational:  device.Rotational,
			Removable:   device.Removable,
			Eligibility: device.Eligibility,
			Reasons:     append([]string(nil), device.Reasons...),
		})
	}
	return DisksResponse{
		CollectedAt: inventory.CollectedAt.UTC().Format(time.RFC3339Nano),
		Status:      inventory.Status,
		Disks:       devices,
	}
}
