package storageinventory

import "time"

const (
	StatusOK          = "ok"
	StatusUnavailable = "unavailable"

	EligibilityEligible  = "eligible"
	EligibilityExcluded  = "excluded"
	EligibilityUnknown   = "unknown"

	ReasonRootDisk              = "root_disk"
	ReasonStateFilesystem       = "state_filesystem"
	ReasonRemovable             = "removable"
	ReasonReadOnly              = "read_only"
	ReasonMountedInUse          = "mounted_in_use"
	ReasonUnsupportedFilesystem = "unsupported_filesystem"
	ReasonInstallMedia          = "install_media"
	ReasonVirtualDevice         = "virtual_device"
	ReasonAmbiguousIdentity     = "ambiguous_identity"
)

// SupportedFSTypes lists filesystems eligible for future pool creation (v1).
var SupportedFSTypes = map[string]struct{}{
	"ext4":  {},
	"xfs":   {},
	"btrfs": {},
}

// Inventory is the platform-level storage discovery result.
type Inventory struct {
	CollectedAt time.Time
	Status      string
	Devices     []Device
}

// Device is a classified block device safe for API projection.
type Device struct {
	ID          string
	Status      string
	SizeBytes   *uint64
	Rotational  *bool
	Removable   *bool
	Eligibility string
	Reasons     []string
}

// RawDevice is internal discovery input before eligibility classification.
type RawDevice struct {
	KernelName        string
	SizeBytes         uint64
	SizeKnown         bool
	Rotational        *bool
	Removable         bool
	ReadOnly          bool
	Virtual           bool
	Optical           bool
	Mounted           bool
	FSTypes           []string
	RootDisk          bool
	StateDisk         bool
	IdentityAmbiguous bool
}
