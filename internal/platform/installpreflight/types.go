package installpreflight

// TargetRequest selects one install target disk explicitly by opaque ID.
type TargetRequest struct {
	DiskID       string
	StateDir     string
	MinSizeBytes uint64
}

// TargetResult is the read-only install-target suitability outcome.
type TargetResult struct {
	DiskID    string
	Status    string
	Reasons   []string
	SizeBytes *uint64
}

// MediaRequest configures optional install-media checks.
type MediaRequest struct {
	EnvelopeRoot        string
	ReleaseManifestPath string
	ArtifactBaseDir     string
}

// MediaResult summarizes install-media and optional release verification.
type MediaResult struct {
	EnvelopeStatus    string
	ReleaseIntegrity  string
	ReleaseAuth       string
	Failures          []CheckFailure
}

// CheckFailure records one failed media or release check.
type CheckFailure struct {
	Scope  string
	Reason string
	Detail string
}

// Result is the combined preflight outcome.
type Result struct {
	Target TargetResult
	Media  MediaResult
}
