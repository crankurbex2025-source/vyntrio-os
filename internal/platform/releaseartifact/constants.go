package releaseartifact

const (
	FormatVersion = "vyntrio-release-manifest-v1"

	IntegrityOK     = "ok"
	IntegrityFailed = "failed"

	AuthenticityNotSigned    = "not_signed"
	AuthenticityUnsupported  = "unsupported"

	ReasonMalformedManifest        = "malformed_manifest"
	ReasonUnsupportedFormatVersion = "unsupported_format_version"
	ReasonUnknownArtifactType      = "unknown_artifact_type"
	ReasonUnknownArtifactUse       = "unknown_artifact_use"
	ReasonDuplicateArtifactName    = "duplicate_artifact_name"
	ReasonInvalidRelativePath      = "invalid_relative_path"
	ReasonMissingFile              = "missing_file"
	ReasonSizeMismatch             = "size_mismatch"
	ReasonSHA256Mismatch           = "sha256_mismatch"
	ReasonInvalidSHA256            = "invalid_sha256"
)

var allowedArtifactTypes = map[string]struct{}{
	"binary":          {},
	"systemd_unit":    {},
	"config_template": {},
	"static_file":     {},
	"archive":         {},
	"iso":             {},
}

var allowedArtifactUses = map[string]struct{}{
	"install_media":        {},
	"appliance_usb_image":  {},
	"recovery_media":       {},
	"release_distribution": {},
	"local_verification":   {},
}

var allowedReleaseChannels = map[string]struct{}{
	"development": {},
	"nightly":     {},
	"beta":        {},
	"stable":      {},
	"hotfix":      {},
}
