package backupstatus

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var allowedFailureClasses = map[string]struct{}{
	FailureArtifact:  {},
	FailureRestart:   {},
	FailureHealth:    {},
	FailureReadiness: {},
	FailureInternal:  {},
}

// EncodeDiskRecord serializes a validated disk record.
func EncodeDiskRecord(record DiskRecord) ([]byte, error) {
	if err := validateDiskRecord(record, time.Now().UTC()); err != nil {
		return nil, err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("backup status: encode failed")
	}
	if len(data) == 0 || len(data) > MaxWriteSize {
		return nil, fmt.Errorf("backup status: encoded size invalid")
	}
	return data, nil
}

// ParseDiskRecord decodes and validates a disk record from JSON bytes.
func ParseDiskRecord(data []byte, now time.Time) (DiskRecord, error) {
	if len(data) == 0 || len(data) > MaxReadSize {
		return DiskRecord{}, fmt.Errorf("backup status: invalid input size")
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return DiskRecord{}, fmt.Errorf("backup status: empty record")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return DiskRecord{}, fmt.Errorf("backup status: malformed json")
	}

	record, err := decodeDiskRecord(raw)
	if err != nil {
		return DiskRecord{}, err
	}
	if err := validateDiskRecord(record, now.UTC()); err != nil {
		return DiskRecord{}, err
	}
	return record, nil
}

func decodeDiskRecord(raw map[string]json.RawMessage) (DiskRecord, error) {
	var record DiskRecord

	schemaVersion, err := parseIntField(raw, "schema_version", true)
	if err != nil {
		return DiskRecord{}, err
	}
	record.SchemaVersion = schemaVersion

	lastOutcome, err := parseStringField(raw, "last_outcome", true)
	if err != nil {
		return DiskRecord{}, err
	}
	record.LastOutcome = lastOutcome

	completedAt, err := parseStringField(raw, "completed_at", true)
	if err != nil {
		return DiskRecord{}, err
	}
	record.CompletedAt = completedAt

	everSucceeded, err := parseBoolField(raw, "ever_succeeded", true)
	if err != nil {
		return DiskRecord{}, err
	}
	record.EverSucceeded = everSucceeded

	switch record.LastOutcome {
	case OutcomeSucceeded:
		if err := requireExactKeys(raw, "schema_version", "last_outcome", "completed_at", "ever_succeeded"); err != nil {
			return DiskRecord{}, err
		}
	case OutcomeFailed:
		if err := requireExactKeys(raw, "schema_version", "last_outcome", "completed_at", "ever_succeeded", "failure_class"); err != nil {
			return DiskRecord{}, err
		}
		failureClass, err := parseStringField(raw, "failure_class", true)
		if err != nil {
			return DiskRecord{}, err
		}
		record.FailureClass = failureClass
	default:
		return DiskRecord{}, fmt.Errorf("backup status: invalid outcome")
	}

	return record, nil
}

func validateDiskRecord(record DiskRecord, now time.Time) error {
	if record.SchemaVersion != SchemaVersion {
		return fmt.Errorf("backup status: unsupported schema")
	}
	switch record.LastOutcome {
	case OutcomeSucceeded:
		if !record.EverSucceeded {
			return fmt.Errorf("backup status: succeeded requires ever_succeeded true")
		}
		if record.FailureClass != "" {
			return fmt.Errorf("backup status: failure_class forbidden on success")
		}
	case OutcomeFailed:
		if record.FailureClass == "" {
			return fmt.Errorf("backup status: missing failure_class")
		}
		if _, ok := allowedFailureClasses[record.FailureClass]; !ok {
			return fmt.Errorf("backup status: invalid failure_class")
		}
	default:
		return fmt.Errorf("backup status: invalid outcome")
	}
	if err := validateTimestamp(record.CompletedAt, now); err != nil {
		return err
	}
	return nil
}

func validateTimestamp(value string, now time.Time) error {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return fmt.Errorf("backup status: invalid timestamp")
	}
	parsed = parsed.UTC()
	if parsed.Before(earliestValidTimestamp) {
		return fmt.Errorf("backup status: timestamp too early")
	}
	if parsed.After(now.Add(timestampSkew)) {
		return fmt.Errorf("backup status: timestamp in future")
	}
	return nil
}

// ProjectDiskRecord maps a validated disk record to the API backup section.
func ProjectDiskRecord(record DiskRecord) Backup {
	completedAt := record.CompletedAt
	ever := record.EverSucceeded
	switch record.LastOutcome {
	case OutcomeSucceeded:
		return Backup{
			Status:        StatusSucceeded,
			CompletedAt:   &completedAt,
			EverSucceeded: &ever,
		}
	case OutcomeFailed:
		failure := record.FailureClass
		return Backup{
			Status:        StatusFailed,
			CompletedAt:   &completedAt,
			EverSucceeded: &ever,
			Failure:       &failure,
		}
	default:
		return Unavailable()
	}
}

// PriorEverSucceeded reports whether a parseable disk record indicates a prior success.
func PriorEverSucceeded(data []byte, now time.Time) bool {
	record, err := ParseDiskRecord(data, now)
	if err != nil {
		return false
	}
	return record.EverSucceeded
}

func requireExactKeys(raw map[string]json.RawMessage, keys ...string) error {
	if len(raw) != len(keys) {
		return fmt.Errorf("backup status: unexpected fields")
	}
	for _, key := range keys {
		if _, ok := raw[key]; !ok {
			return fmt.Errorf("backup status: missing field")
		}
	}
	for key := range raw {
		found := false
		for _, allowed := range keys {
			if key == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("backup status: unknown field")
		}
	}
	return nil
}

func parseStringField(raw map[string]json.RawMessage, key string, required bool) (string, error) {
	valueRaw, ok := raw[key]
	if !ok {
		if required {
			return "", fmt.Errorf("backup status: missing field")
		}
		return "", nil
	}
	var value string
	if err := json.Unmarshal(valueRaw, &value); err != nil || value == "" {
		return "", fmt.Errorf("backup status: invalid field")
	}
	return value, nil
}

func parseIntField(raw map[string]json.RawMessage, key string, required bool) (int, error) {
	valueRaw, ok := raw[key]
	if !ok {
		if required {
			return 0, fmt.Errorf("backup status: missing field")
		}
		return 0, nil
	}
	var value int
	if err := json.Unmarshal(valueRaw, &value); err != nil {
		return 0, fmt.Errorf("backup status: invalid field")
	}
	return value, nil
}

func parseBoolField(raw map[string]json.RawMessage, key string, required bool) (bool, error) {
	valueRaw, ok := raw[key]
	if !ok {
		if required {
			return false, fmt.Errorf("backup status: missing field")
		}
		return false, nil
	}
	var value bool
	if err := json.Unmarshal(valueRaw, &value); err != nil {
		return false, fmt.Errorf("backup status: invalid field")
	}
	return value, nil
}

// BuildSucceededRecord constructs a succeeded disk record.
func BuildSucceededRecord(completedAt time.Time) DiskRecord {
	return DiskRecord{
		SchemaVersion: SchemaVersion,
		LastOutcome:   OutcomeSucceeded,
		CompletedAt:   completedAt.UTC().Format(time.RFC3339Nano),
		EverSucceeded: true,
	}
}

// BuildFailedRecord constructs a failed disk record preserving prior success.
func BuildFailedRecord(completedAt time.Time, failureClass string, priorEverSucceeded bool) DiskRecord {
	return DiskRecord{
		SchemaVersion: SchemaVersion,
		LastOutcome:   OutcomeFailed,
		CompletedAt:   completedAt.UTC().Format(time.RFC3339Nano),
		EverSucceeded: priorEverSucceeded,
		FailureClass:  failureClass,
	}
}
