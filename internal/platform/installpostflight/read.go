package installpostflight

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadRecord loads HANDOVER_RECORD.txt from a sandbox target root.
func ReadRecord(targetRoot string) (Handover, error) {
	path := filepath.Join(targetRoot, HandoverRecordName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Handover{}, err
	}
	return ParseRecord(string(data))
}

// ParseRecord parses a handover record file.
func ParseRecord(content string) (Handover, error) {
	h := Handover{DeferredItems: []string{}, PayloadPaths: []string{}, NextSteps: []string{}}
	scanner := bufio.NewScanner(strings.NewReader(content))
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "deferred_items:") {
			section = "deferred"
			continue
		}
		if strings.HasPrefix(line, "applied_payloads:") {
			section = "payloads"
			continue
		}
		if strings.HasPrefix(line, "next_steps:") {
			section = "next"
			continue
		}
		if strings.HasPrefix(line, "  - ") {
			item := strings.TrimPrefix(line, "  - ")
			switch section {
			case "deferred":
				h.DeferredItems = append(h.DeferredItems, item)
			case "payloads":
				h.PayloadPaths = append(h.PayloadPaths, strings.TrimPrefix(item, "/"))
			case "next":
				h.NextSteps = append(h.NextSteps, item)
			}
			continue
		}
		section = ""
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "schema_version":
			h.SchemaVersion = value
		case "command":
			h.Command = value
		case "overall_status":
			h.OverallStatus = value
		case "target_disk_id":
			h.TargetDiskID = value
		case "release_manifest":
			h.ReleaseManifestPath = value
		case "release_version":
			h.ReleaseVersion = value
		case "envelope_root":
			h.EnvelopeRoot = value
		case "preflight_status":
			h.PreflightStatus = value
		case "target_preflight":
			h.TargetPreflight = value
		case "media_envelope":
			h.MediaEnvelope = value
		case "release_integrity":
			h.ReleaseIntegrity = value
		case "release_authenticity":
			h.ReleaseAuthenticity = value
		case "install_write":
			h.InstallWrite = value
		case "sandbox_target_path":
			h.SandboxTargetPath = value
		case "payloads_copied":
			h.PayloadsCopied, _ = strconv.Atoi(value)
		case "payload_allowlist":
			h.PayloadAllowlist, _ = strconv.Atoi(value)
		case "allowlist_complete":
			h.AllowlistComplete = value == "true"
		case "mutation_scope":
			h.MutationScope = value
		case "failure_stage":
			h.FailureStage = value
		case "failure_reason":
			h.FailureReason = value
		}
	}
	if err := scanner.Err(); err != nil {
		return Handover{}, err
	}
	if h.SchemaVersion == "" {
		return Handover{}, os.ErrNotExist
	}
	return h, nil
}
