package installpreflight

const (
	DefaultStateDir = "/var/lib/vyntrio"

	// MinInstallSizeBytes is the default minimum target size for greenfield install planning.
	MinInstallSizeBytes uint64 = 8 * 1024 * 1024 * 1024

	TargetEligible  = "eligible"
	TargetExcluded  = "excluded"
	TargetUnknown   = "unknown"
	TargetNotFound  = "not_found"
	TargetAmbiguous = "ambiguous"

	MediaOK          = "ok"
	MediaSkipped     = "skipped"
	MediaFailed      = "failed"
	ReleaseSkipped   = "skipped"
	ReleaseOK        = "ok"
	ReleaseFailed    = "failed"

	ReasonTargetSelectionRequired = "target_selection_required"
	ReasonTargetNotFound          = "target_not_found"
	ReasonTargetAmbiguous         = "target_ambiguous"
	ReasonInsufficientSize        = "insufficient_size"
	ReasonDiscoveryUnavailable    = "discovery_unavailable"

	ReasonEnvelopeMissing         = "envelope_missing"
	ReasonEnvelopeRecordMissing   = "envelope_record_missing"
	ReasonEnvelopeRoleMismatch    = "envelope_role_mismatch"
	ReasonPayloadRootMissing      = "payload_root_missing"
	ReasonPayloadFileMissing      = "payload_file_missing"
	ReasonPayloadInventoryExtra   = "payload_inventory_extra"
	ReasonPayloadInventoryMissing = "payload_inventory_missing"
	ReasonPayloadExcludedPattern  = "payload_excluded_pattern"
	ReasonPayloadExcludedBasename = "payload_excluded_basename"
)

var requiredPayloadFiles = []string{
	"usr/bin/vyntrio-api",
	"usr/bin/vyntrio-backup",
	"etc/systemd/system/vyntrio-api.service",
	"usr/lib/sysusers.d/vyntrio.conf",
	"etc/tmpfiles.d/vyntrio.conf",
	"etc/vyntrio/config.toml",
}
