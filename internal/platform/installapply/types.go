package installapply

import (
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
)

// Request configures an existing-partition payload apply.
type Request struct {
	TargetDiskID    string
	StateDir        string
	EnvelopeRoot    string
	ReleaseManifest string
	ArtifactBaseDir string
	MountRoot       string
	Force           bool
}

// Outcome records partition apply execution facts.
type Outcome struct {
	TargetDiskID           string
	MountPoint             string
	TargetFSType           string
	PartitionDevicePath    string
	PayloadsCopied         int
	PayloadPaths           []string
	ReleaseVersion         string
	Preflight              installpreflight.Result
	PreflightPassed        bool
	HostBlockDeviceMutated bool
	ApplyRecordPath        string
	MutationRecordPath     string
	FailureStage           string
	FailureReason          string
}
