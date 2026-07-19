package installwrite

import "os"
type Request struct {
	TargetDiskID      string
	TargetRoot        string
	SandboxRoot       string
	StateDir          string
	EnvelopeRoot      string
	ReleaseManifest   string
	ArtifactBaseDir   string
	Force             bool
	ApplyTarget       bool
	MountRoot         string
}

// Result summarizes a completed install write.
type Result struct {
	TargetDiskID   string
	TargetRoot     string
	PayloadsCopied int
	ReleaseVersion string
}

// CopyEntry is one payload file copy operation.
type CopyEntry struct {
	SourcePath     string
	TargetRel      string
	Mode           os.FileMode
	ExpectedSHA256 string
	ExpectedSize   uint64
}
