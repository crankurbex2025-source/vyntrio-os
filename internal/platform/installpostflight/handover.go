package installpostflight

import "time"

// Input captures execution facts for handover generation without re-running checks.
type Input struct {
	Command             string
	TargetDiskID        string
	ReleaseManifestPath string
	ReleaseVersion      string
	EnvelopeRoot        string
	TargetPreflight     string
	MediaEnvelope       string
	ReleaseIntegrity    string
	ReleaseAuthenticity string
	PreflightOK         bool
	InstallWrite        string
	SandboxTargetPath   string
	PayloadsCopied      int
	PayloadAllowlist    int
	PayloadPaths        []string
	FailureStage        string
	FailureReason       string
	ApplyTarget         bool
	HostBlockDeviceMutated bool
	TargetMutationRecord   string
	PartitionApplyRecord   string
	MountPoint             string
	PartitionDevicePath    string
	GeneratedAt         time.Time
}

// Handover is the operator-facing postflight summary and record.
type Handover struct {
	SchemaVersion       string
	GeneratedAt         time.Time
	Command             string
	OverallStatus       string
	TargetDiskID        string
	ReleaseManifestPath string
	ReleaseVersion      string
	EnvelopeRoot        string
	PreflightStatus     string
	TargetPreflight     string
	MediaEnvelope       string
	ReleaseIntegrity    string
	ReleaseAuthenticity string
	InstallWrite        string
	SandboxTargetPath   string
	PayloadsCopied      int
	PayloadAllowlist    int
	AllowlistComplete   bool
	PayloadPaths        []string
	DeferredItems       []string
	MutationScope       string
	FailureStage        string
	FailureReason       string
	ApplyTarget         bool
	HostBlockDeviceMutated bool
	TargetMutationRecord   string
	PartitionApplyRecord   string
	MountPoint             string
	PartitionDevicePath    string
	NextSteps           []string
}

// Build assembles a handover from recorded execution facts.
func Build(in Input) Handover {
	if in.GeneratedAt.IsZero() {
		in.GeneratedAt = time.Now().UTC()
	}
	allowlist := in.PayloadAllowlist
	if allowlist == 0 {
		allowlist = 6
	}

	h := Handover{
		SchemaVersion:       SchemaVersion,
		GeneratedAt:         in.GeneratedAt,
		Command:             in.Command,
		TargetDiskID:        in.TargetDiskID,
		ReleaseManifestPath: in.ReleaseManifestPath,
		ReleaseVersion:      in.ReleaseVersion,
		EnvelopeRoot:        in.EnvelopeRoot,
		TargetPreflight:     in.TargetPreflight,
		MediaEnvelope:       in.MediaEnvelope,
		ReleaseIntegrity:    in.ReleaseIntegrity,
		ReleaseAuthenticity: in.ReleaseAuthenticity,
		InstallWrite:        in.InstallWrite,
		SandboxTargetPath:   in.SandboxTargetPath,
		PayloadsCopied:      in.PayloadsCopied,
		PayloadAllowlist:    allowlist,
		AllowlistComplete:   in.PayloadsCopied > 0 && in.PayloadsCopied == allowlist,
		PayloadPaths:        append([]string(nil), in.PayloadPaths...),
		DeferredItems:       append([]string(nil), defaultDeferredItems...),
		FailureStage:        in.FailureStage,
		FailureReason:       in.FailureReason,
		ApplyTarget:         in.ApplyTarget,
		HostBlockDeviceMutated: in.HostBlockDeviceMutated,
		TargetMutationRecord:   in.TargetMutationRecord,
		PartitionApplyRecord:   in.PartitionApplyRecord,
		MountPoint:             in.MountPoint,
		PartitionDevicePath:      in.PartitionDevicePath,
	}
	if in.ApplyTarget {
		h.MutationScope = MutationScopeTargetApply
	} else {
		h.MutationScope = MutationScopeStatement
	}

	switch in.Command {
	case CommandPreflight:
		h.InstallWrite = WriteNotRun
		if in.PreflightOK {
			h.PreflightStatus = PreflightOK
			h.OverallStatus = OverallPreflightOnly
			h.NextSteps = append([]string(nil), defaultNextStepsAfterPreflight...)
		} else {
			h.PreflightStatus = PreflightFailed
			h.OverallStatus = OverallFailed
			h.NextSteps = append([]string(nil), defaultNextStepsAfterFailure...)
		}
	case CommandInstall, CommandApply:
		if in.PreflightOK {
			h.PreflightStatus = PreflightOK
		} else {
			h.PreflightStatus = PreflightFailed
		}
		switch in.InstallWrite {
		case WriteSucceeded:
			h.OverallStatus = OverallSucceeded
			h.NextSteps = append([]string(nil), defaultNextStepsAfterSuccess...)
		case WritePartial:
			h.OverallStatus = OverallFailed
			h.NextSteps = append([]string(nil), defaultNextStepsAfterFailure...)
		default:
			h.OverallStatus = OverallFailed
			h.NextSteps = append([]string(nil), defaultNextStepsAfterFailure...)
		}
	default:
		h.OverallStatus = OverallFailed
	}

	if in.PreflightOK && in.MediaEnvelope == "" {
		h.MediaEnvelope = PreflightSkipped
	}
	if in.PreflightOK && in.ReleaseIntegrity == "" {
		h.ReleaseIntegrity = PreflightSkipped
	}
	return h
}
