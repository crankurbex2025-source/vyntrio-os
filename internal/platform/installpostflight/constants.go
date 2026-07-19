package installpostflight

const (
	SchemaVersion = "vyntrio-installer-handover-v1"
	HandoverRecordName = "HANDOVER_RECORD.txt"

	CommandPreflight = "preflight"
	CommandInstall   = "install"
	CommandApply     = "apply"
	CommandPostflight = "postflight"

	OverallSucceeded     = "succeeded"
	OverallFailed        = "failed"
	OverallPreflightOnly = "preflight_only"

	PreflightOK      = "ok"
	PreflightFailed  = "failed"
	PreflightSkipped = "skipped"

	WriteNotRun    = "not_run"
	WriteSucceeded = "succeeded"
	WriteFailed    = "failed"
	WritePartial   = "partial"

	MutationScopeStatement       = "sandbox_payload_copy_only; no real block-device mutation"
	MutationScopeTargetApply     = "mounted_partition_payload_copy_only; six allowlisted payloads; not full OS install"
)

var defaultDeferredItems = []string{
	"block_device_partitioning",
	"block_device_formatting",
	"host_block_device_write",
	"service_enablement",
	"service_start",
	"sysusers_tmpfiles_application",
	"bootstrap_handoff",
	"full_os_install",
}

var defaultNextStepsAfterSuccess = []string{
	"Review HANDOVER_RECORD.txt and INSTALL_RECORD.txt in the sandbox target tree.",
	"Do not treat sandbox payload copy as a completed hardware installation.",
	"Future slices must enable services and bootstrap before appliance use.",
}

var defaultNextStepsAfterPreflight = []string{
	"Run install with --force only after reviewing preflight results.",
	"Obtain opaque disk ID from GET /api/v1/storage/disks if needed.",
}

var defaultNextStepsAfterFailure = []string{
	"Resolve the reported failure stage before retrying.",
	"Do not assume any host block device was modified.",
}
