package storageinventory

import (
	"context"
	"strings"
	"time"
)

// BlockDeviceReader discovers raw block devices for classification.
type BlockDeviceReader interface {
	ListBlockDevices(stateDir string) ([]RawDevice, error)
}

// CollectorDeps configures injectable discovery readers for tests.
type CollectorDeps struct {
	Reader BlockDeviceReader
	Clock  func() int64 // unix seconds; optional for tests
}

// Collector assembles read-only storage inventory.
type Collector struct {
	stateDir string
	reader   BlockDeviceReader
}

// NewCollector creates a storage inventory collector.
func NewCollector(stateDir string, deps CollectorDeps) Collector {
	reader := deps.Reader
	if reader == nil {
		reader = defaultBlockDeviceReader()
	}
	return Collector{
		stateDir: stateDir,
		reader:   reader,
	}
}

// CollectRaw returns unclassified block devices from the configured reader.
func (c Collector) CollectRaw(stateDir string) ([]RawDevice, error) {
	if strings.TrimSpace(stateDir) == "" {
		stateDir = c.stateDir
	}
	return c.reader.ListBlockDevices(stateDir)
}

// Collect returns classified storage inventory with safe degradation.
func (c Collector) Collect(ctx context.Context) Inventory {
	_ = ctx
	now := currentTime()
	rawDevices, err := c.reader.ListBlockDevices(c.stateDir)
	if err != nil {
		return Inventory{
			CollectedAt: now,
			Status:      StatusUnavailable,
			Devices:     nil,
		}
	}

	devices := make([]Device, 0, len(rawDevices))
	for _, raw := range rawDevices {
		devices = append(devices, Classify(raw))
	}
	return Inventory{
		CollectedAt: now,
		Status:      StatusOK,
		Devices:     devices,
	}
}

func currentTime() time.Time {
	return time.Now().UTC()
}
