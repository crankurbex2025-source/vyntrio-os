package storageinventory

import "sort"

// Classify applies fail-closed eligibility rules to a discovered device.
func Classify(raw RawDevice) Device {
	id, ok := stableDeviceID(raw.KernelName)
	if !ok || raw.IdentityAmbiguous {
		return Device{
			ID:          "disk-unknown",
			Status:      StatusUnavailable,
			Eligibility: EligibilityUnknown,
			Reasons:     []string{ReasonAmbiguousIdentity},
		}
	}

	reasons := exclusionReasons(raw)
	sort.Strings(reasons)

	device := Device{
		ID:     id,
		Status: StatusOK,
	}
	if raw.SizeKnown {
		size := raw.SizeBytes
		device.SizeBytes = &size
	}
	if raw.Rotational != nil {
		rot := *raw.Rotational
		device.Rotational = &rot
	}
	rem := raw.Removable
	device.Removable = &rem

	switch {
	case !raw.SizeKnown || raw.SizeBytes == 0:
		device.Eligibility = EligibilityUnknown
		device.Reasons = uniqueReasons(append(reasons, ReasonAmbiguousIdentity))
	case len(reasons) > 0:
		device.Eligibility = EligibilityExcluded
		device.Reasons = reasons
	default:
		device.Eligibility = EligibilityEligible
	}
	return device
}

func exclusionReasons(raw RawDevice) []string {
	var reasons []string
	if raw.Virtual {
		reasons = append(reasons, ReasonVirtualDevice)
	}
	if raw.Optical {
		reasons = append(reasons, ReasonInstallMedia)
	}
	if raw.RootDisk {
		reasons = append(reasons, ReasonRootDisk)
	}
	if raw.StateDisk {
		reasons = append(reasons, ReasonStateFilesystem)
	}
	if raw.Removable {
		reasons = append(reasons, ReasonRemovable)
	}
	if raw.ReadOnly {
		reasons = append(reasons, ReasonReadOnly)
	}
	if raw.Mounted && !raw.RootDisk && !raw.StateDisk {
		reasons = append(reasons, ReasonMountedInUse)
	}
	if hasUnsupportedFilesystem(raw.FSTypes) {
		reasons = append(reasons, ReasonUnsupportedFilesystem)
	}
	return uniqueReasons(reasons)
}

func hasUnsupportedFilesystem(fsTypes []string) bool {
	if len(fsTypes) == 0 {
		return false
	}
	for _, fsType := range fsTypes {
		if fsType == "" {
			continue
		}
		if _, ok := SupportedFSTypes[fsType]; !ok {
			return true
		}
	}
	return false
}

func uniqueReasons(reasons []string) []string {
	if len(reasons) == 0 {
		return nil
	}
	sort.Strings(reasons)
	out := reasons[:1]
	for _, reason := range reasons[1:] {
		if reason != out[len(out)-1] {
			out = append(out, reason)
		}
	}
	return out
}
