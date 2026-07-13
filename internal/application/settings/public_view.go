package settings

// PublicSettingsResponse is the safe owner-visible settings view for GET /api/v1/settings.
type PublicSettingsResponse struct {
	Instance PublicInstanceSettings `json:"instance"`
	API      PublicAPISettings      `json:"api"`
}

// PublicInstanceSettings exposes non-secret instance metadata.
type PublicInstanceSettings struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PublicAPISettings exposes non-secret API runtime metadata.
type PublicAPISettings struct {
	Environment string `json:"environment"`
}

// PublicView assembles the read-only public settings response from startup snapshot and config.
type PublicView struct {
	snapshot    Snapshot
	version     string
	environment string
}

// NewPublicView creates a public settings view from validated startup state and config metadata.
func NewPublicView(snapshot Snapshot, version, environment string) PublicView {
	return PublicView{
		snapshot:    snapshot,
		version:     version,
		environment: environment,
	}
}

// Response returns the explicit safe DTO for JSON serialization.
func (v PublicView) Response() PublicSettingsResponse {
	return PublicSettingsResponse{
		Instance: PublicInstanceSettings{
			Name:    v.snapshot.Hostname,
			Version: v.version,
		},
		API: PublicAPISettings{
			Environment: v.environment,
		},
	}
}
