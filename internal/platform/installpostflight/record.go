package installpostflight

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteRecord persists HANDOVER_RECORD.txt under the sandbox target root.
func WriteRecord(targetRoot string, handover Handover) (string, error) {
	if strings.TrimSpace(targetRoot) == "" {
		return "", fmt.Errorf("handover record requires target root")
	}
	path := filepath.Join(targetRoot, HandoverRecordName)
	if err := os.WriteFile(path, EncodeRecord(handover), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// EncodeRecord renders the handover artifact.
func EncodeRecord(h Handover) []byte {
	if h.GeneratedAt.IsZero() {
		h.GeneratedAt = time.Now().UTC()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Vyntrio installer handover record — do not treat as full OS install\n")
	fmt.Fprintf(&b, "schema_version: %s\n", h.SchemaVersion)
	fmt.Fprintf(&b, "generated_at: %s\n", h.GeneratedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "command: %s\n", h.Command)
	fmt.Fprintf(&b, "overall_status: %s\n", h.OverallStatus)
	fmt.Fprintf(&b, "target_disk_id: %s\n", h.TargetDiskID)
	if h.ReleaseManifestPath != "" {
		fmt.Fprintf(&b, "release_manifest: %s\n", h.ReleaseManifestPath)
	}
	if h.ReleaseVersion != "" {
		fmt.Fprintf(&b, "release_version: %s\n", h.ReleaseVersion)
	}
	if h.EnvelopeRoot != "" {
		fmt.Fprintf(&b, "envelope_root: %s\n", h.EnvelopeRoot)
	}
	fmt.Fprintf(&b, "preflight_status: %s\n", h.PreflightStatus)
	fmt.Fprintf(&b, "target_preflight: %s\n", valueOrUnknown(h.TargetPreflight))
	fmt.Fprintf(&b, "media_envelope: %s\n", valueOrUnknown(h.MediaEnvelope))
	fmt.Fprintf(&b, "release_integrity: %s\n", valueOrUnknown(h.ReleaseIntegrity))
	if h.ReleaseAuthenticity != "" {
		fmt.Fprintf(&b, "release_authenticity: %s\n", h.ReleaseAuthenticity)
	}
	fmt.Fprintf(&b, "install_write: %s\n", h.InstallWrite)
	if h.SandboxTargetPath != "" {
		fmt.Fprintf(&b, "sandbox_target_path: %s\n", h.SandboxTargetPath)
	}
	fmt.Fprintf(&b, "payloads_copied: %d\n", h.PayloadsCopied)
	fmt.Fprintf(&b, "payload_allowlist: %d\n", h.PayloadAllowlist)
	fmt.Fprintf(&b, "allowlist_complete: %t\n", h.AllowlistComplete)
	fmt.Fprintf(&b, "mutation_scope: %s\n", h.MutationScope)
	fmt.Fprintf(&b, "apply_target: %t\n", h.ApplyTarget)
	fmt.Fprintf(&b, "host_block_device_mutated: %t\n", h.HostBlockDeviceMutated)
	if h.TargetMutationRecord != "" {
		fmt.Fprintf(&b, "target_mutation_record: %s\n", h.TargetMutationRecord)
	}
	if h.PartitionApplyRecord != "" {
		fmt.Fprintf(&b, "partition_apply_record: %s\n", h.PartitionApplyRecord)
	}
	if h.PartitionDevicePath != "" {
		fmt.Fprintf(&b, "partition_device_path: %s\n", h.PartitionDevicePath)
	}
	if h.MountPoint != "" {
		fmt.Fprintf(&b, "mount_point: %s\n", h.MountPoint)
	}
	if h.FailureStage != "" {
		fmt.Fprintf(&b, "failure_stage: %s\n", h.FailureStage)
	}
	if h.FailureReason != "" {
		fmt.Fprintf(&b, "failure_reason: %s\n", h.FailureReason)
	}
	fmt.Fprintf(&b, "deferred_items:\n")
	for _, item := range h.DeferredItems {
		fmt.Fprintf(&b, "  - %s\n", item)
	}
	if len(h.PayloadPaths) > 0 {
		fmt.Fprintf(&b, "applied_payloads:\n")
		for _, path := range h.PayloadPaths {
			fmt.Fprintf(&b, "  - /%s\n", path)
		}
	}
	fmt.Fprintf(&b, "next_steps:\n")
	for _, step := range h.NextSteps {
		fmt.Fprintf(&b, "  - %s\n", step)
	}
	return []byte(b.String())
}

// SummaryLine returns a concise stdout line for operators.
func SummaryLine(h Handover) string {
	return fmt.Sprintf(
		"installer postflight: overall=%s command=%s preflight=%s install_write=%s disk_id=%s sandbox=%s payloads=%d/%d mutation_scope=%s",
		h.OverallStatus,
		h.Command,
		h.PreflightStatus,
		h.InstallWrite,
		h.TargetDiskID,
		h.SandboxTargetPath,
		h.PayloadsCopied,
		h.PayloadAllowlist,
		h.MutationScope,
	)
}

// NoteLines returns stderr guidance lines.
func NoteLines(h Handover) []string {
	prefix := "sandbox-only staged install"
	if h.ApplyTarget {
		prefix = "guarded target mutation; not a completed hardware/OS installation"
	}
	lines := []string{
		"installer postflight note: " + prefix,
		fmt.Sprintf("installer postflight note: %s", h.MutationScope),
	}
	if h.OverallStatus == OverallFailed {
		if h.FailureStage != "" {
			lines = append(lines, fmt.Sprintf("installer postflight note: failure_stage=%s", h.FailureStage))
		}
		if h.FailureReason != "" {
			lines = append(lines, fmt.Sprintf("installer postflight note: failure_reason=%s", h.FailureReason))
		}
	}
	for _, step := range h.NextSteps {
		lines = append(lines, "installer postflight next: "+step)
	}
	return lines
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
