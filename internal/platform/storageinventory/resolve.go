package storageinventory

import "fmt"

// LookupRawDevice finds the discovered raw device for an opaque disk ID.
// Intended for privileged installer flows only; kernel names must not appear in APIs.
func LookupRawDevice(devices []RawDevice, diskID string) (RawDevice, int, error) {
	var found RawDevice
	count := 0
	for _, device := range devices {
		id, ok := stableDeviceID(device.KernelName)
		if !ok {
			continue
		}
		if id != diskID {
			continue
		}
		found = device
		count++
	}
	switch count {
	case 0:
		return RawDevice{}, 0, fmt.Errorf("device not found")
	case 1:
		return found, 1, nil
	default:
		return RawDevice{}, count, fmt.Errorf("ambiguous device identity")
	}
}
