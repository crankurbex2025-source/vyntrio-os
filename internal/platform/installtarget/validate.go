package installtarget

import (
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

func validateDevice(device RawDeviceInput, diskID string) error {
	if device.IdentityAmbiguous || !device.SizeKnown || device.SizeBytes == 0 {
		return ErrUnsupportedTargetState
	}
	if device.RootDisk || device.StateDisk {
		return ErrTargetNotEligible
	}
	if device.Removable || device.Virtual || device.Optical || device.ReadOnly {
		return ErrTargetNotEligible
	}
	if device.Mounted {
		return ErrTargetMounted
	}

	classified := storageinventory.Classify(rawFromInput(device))
	if classified.ID != diskID {
		return ErrUnsupportedTargetState
	}
	if classified.Eligibility != storageinventory.EligibilityEligible {
		return ErrTargetNotEligible
	}
	return nil
}

func lookupDevice(reader DeviceReader, stateDir, diskID string) (RawDeviceInput, error) {
	devices, err := reader.ListBlockDevices(stateDir)
	if err != nil {
		return RawDeviceInput{}, err
	}
	rawDevices := make([]storageinventory.RawDevice, 0, len(devices))
	for _, device := range devices {
		rawDevices = append(rawDevices, rawFromInput(device))
	}
	raw, count, err := storageinventory.LookupRawDevice(rawDevices, diskID)
	if err != nil {
		if count > 1 {
			return RawDeviceInput{}, ErrAmbiguousTargetLayout
		}
		return RawDeviceInput{}, ErrUnsupportedTargetState
	}
	return rawDeviceInputFromRaw(raw), nil
}

func rawFromInput(device RawDeviceInput) storageinventory.RawDevice {
	return storageinventory.RawDevice{
		KernelName:        device.KernelName,
		SizeBytes:         device.SizeBytes,
		SizeKnown:         device.SizeKnown,
		Removable:         device.Removable,
		ReadOnly:          device.ReadOnly,
		Virtual:           device.Virtual,
		Optical:           device.Optical,
		Mounted:           device.Mounted,
		RootDisk:          device.RootDisk,
		StateDisk:         device.StateDisk,
		IdentityAmbiguous: device.IdentityAmbiguous,
	}
}

func mountPointFor(diskID, mountRoot string) string {
	root := strings.TrimSpace(mountRoot)
	if root == "" {
		root = DefaultMountRoot
	}
	return strings.TrimRight(root, "/") + "/" + strings.TrimSpace(diskID)
}
