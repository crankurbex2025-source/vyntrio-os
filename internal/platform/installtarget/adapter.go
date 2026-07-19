package installtarget

import (
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

// InventoryAdapter resolves opaque disk IDs via the storage inventory collector.
type InventoryAdapter struct {
	Collector storageinventory.Collector
}

// NewInventoryAdapter creates a device reader backed by storage inventory.
func NewInventoryAdapter(stateDir string) InventoryAdapter {
	return InventoryAdapter{
		Collector: storageinventory.NewCollector(stateDir, storageinventory.CollectorDeps{}),
	}
}

// ListBlockDevices returns raw discovery inputs for target validation.
func (a InventoryAdapter) ListBlockDevices(stateDir string) ([]RawDeviceInput, error) {
	rawDevices, err := a.Collector.CollectRaw(stateDir)
	if err != nil {
		return nil, err
	}
	out := make([]RawDeviceInput, 0, len(rawDevices))
	for _, raw := range rawDevices {
		out = append(out, rawDeviceInputFromRaw(raw))
	}
	return out, nil
}

// IsMounted reports whether a kernel block device name is currently mounted.
func (a InventoryAdapter) IsMounted(kernelName string, stateDir string) (bool, error) {
	rawDevices, err := a.Collector.CollectRaw(stateDir)
	if err != nil {
		return false, err
	}
	for _, raw := range rawDevices {
		if raw.KernelName != kernelName {
			continue
		}
		return raw.Mounted, nil
	}
	return false, nil
}

func rawDeviceInputFromRaw(raw storageinventory.RawDevice) RawDeviceInput {
	return RawDeviceInput{
		KernelName:        raw.KernelName,
		SizeBytes:         raw.SizeBytes,
		SizeKnown:         raw.SizeKnown,
		Removable:         raw.Removable,
		ReadOnly:          raw.ReadOnly,
		Virtual:           raw.Virtual,
		Optical:           raw.Optical,
		Mounted:           raw.Mounted,
		RootDisk:          raw.RootDisk,
		StateDisk:         raw.StateDisk,
		IdentityAmbiguous: raw.IdentityAmbiguous,
	}
}

// StubReader is an injectable device reader for tests.
type StubReader struct {
	Devices []RawDeviceInput
	Err     error
}

func (s StubReader) ListBlockDevices(string) ([]RawDeviceInput, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return append([]RawDeviceInput(nil), s.Devices...), nil
}

func (s StubReader) IsMounted(kernelName string, _ string) (bool, error) {
	for _, device := range s.Devices {
		if device.KernelName == kernelName {
			return device.Mounted, nil
		}
	}
	return false, nil
}
