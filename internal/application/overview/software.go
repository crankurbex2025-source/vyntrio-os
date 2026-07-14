package overview

import "strings"

const (
	SoftwareStatusOK          = "ok"
	SoftwareStatusUnavailable = "unavailable"

	ReleaseChannelDevelopment = "development"
	ReleaseChannelProduction  = "production"
	ReleaseChannelUnknown     = "unknown"
)

// SoftwareSection exposes read-only application release metadata for the overview.
type SoftwareSection struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
	Channel string `json:"channel,omitempty"`
}

// AssembleSoftware maps already-local loader metadata into the safe software section.
func AssembleSoftware(version, commit, environment string) SoftwareSection {
	version = strings.TrimSpace(version)
	if version == "" {
		return SoftwareSection{Status: SoftwareStatusUnavailable}
	}

	section := SoftwareSection{
		Status:  SoftwareStatusOK,
		Version: version,
		Channel: MapReleaseChannel(environment),
	}
	if commit = strings.TrimSpace(commit); commit != "" {
		section.Commit = commit
	}
	return section
}

// MapReleaseChannel derives the coarse release channel from the existing API environment field.
func MapReleaseChannel(environment string) string {
	switch strings.TrimSpace(environment) {
	case ReleaseChannelDevelopment, ReleaseChannelProduction:
		return environment
	default:
		return ReleaseChannelUnknown
	}
}
