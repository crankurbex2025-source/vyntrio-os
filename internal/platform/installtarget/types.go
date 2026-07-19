package installtarget

// MountCandidate is one supported filesystem source on the selected device.
type MountCandidate struct {
	DevicePath string
	FSType     string
}

// ApplyRequest configures a guarded target mutation apply.
type ApplyRequest struct {
	DiskID    string
	StateDir  string
	MountRoot string
	Force     bool
}

// ApplyOutcome records what happened during target mutation.
type ApplyOutcome struct {
	DiskID         string
	MountPoint     string
	FSType         string
	Candidate      MountCandidate
	Mounted        bool
	RolledBack     bool
	RecordPath     string
	FailureReason  string
	FailureStage   string
}

// MountController performs mount and unmount operations.
type MountController interface {
	Mount(devicePath, mountPoint, fsType string) error
	Unmount(mountPoint string) error
}

// FSProber detects filesystem type for a block device path.
type FSProber interface {
	ProbeFSType(devicePath string) (string, error)
}

// DeviceReader supplies raw devices for resolution.
type DeviceReader interface {
	ListBlockDevices(stateDir string) ([]RawDeviceInput, error)
	IsMounted(kernelName string, stateDir string) (bool, error)
}

// RawDeviceInput mirrors discovery fields needed for target validation.
type RawDeviceInput struct {
	KernelName        string
	SizeBytes         uint64
	SizeKnown         bool
	Removable         bool
	ReadOnly          bool
	Virtual           bool
	Optical           bool
	Mounted           bool
	RootDisk          bool
	StateDisk         bool
	IdentityAmbiguous bool
}
