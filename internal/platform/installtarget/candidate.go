package installtarget

import "fmt"

// PartitionLister returns partition kernel names for a parent block device.
type PartitionLister interface {
	ListPartitions(kernelName string) ([]string, error)
}

// StubPartitionLister is an injectable partition lister for tests.
type StubPartitionLister struct {
	Partitions map[string][]string
	Err        error
}

func (s StubPartitionLister) ListPartitions(kernelName string) ([]string, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	if s.Partitions == nil {
		return nil, nil
	}
	return append([]string(nil), s.Partitions[kernelName]...), nil
}

func resolveMountCandidate(
	kernelName string,
	lister PartitionLister,
	prober FSProber,
	reader DeviceReader,
	stateDir string,
) (MountCandidate, error) {
	partitions, err := lister.ListPartitions(kernelName)
	if err != nil {
		return MountCandidate{}, fmt.Errorf("%w: list partitions: %v", ErrUnsupportedTargetState, err)
	}

	devicePaths := make([]string, 0, len(partitions)+1)
	if len(partitions) == 0 {
		devicePaths = append(devicePaths, devicePath(kernelName))
	} else {
		for _, part := range partitions {
			devicePaths = append(devicePaths, devicePath(part))
		}
	}

	var candidates []MountCandidate
	for i, devicePath := range devicePaths {
		partName := kernelName
		if len(partitions) > 0 {
			partName = partitions[i]
		}
		mounted, err := reader.IsMounted(partName, stateDir)
		if err != nil {
			return MountCandidate{}, fmt.Errorf("%w: mount state: %v", ErrUnsupportedTargetState, err)
		}
		if mounted {
			continue
		}

		fsType, err := prober.ProbeFSType(devicePath)
		if err != nil || fsType == "" {
			continue
		}
		if _, ok := supportedFSTypes[fsType]; !ok {
			continue
		}
		candidates = append(candidates, MountCandidate{
			DevicePath: devicePath,
			FSType:     fsType,
		})
	}

	switch len(candidates) {
	case 0:
		return MountCandidate{}, ErrUnsupportedTargetState
	case 1:
		return candidates[0], nil
	default:
		return MountCandidate{}, ErrAmbiguousTargetLayout
	}
}

func devicePath(kernelName string) string {
	return "/dev/" + kernelName
}
