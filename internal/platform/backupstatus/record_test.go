package backupstatus_test

import (
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

func fixedNow() time.Time {
	return time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
}

func TestParseDiskRecordSucceeded(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"last_outcome": "succeeded",
		"completed_at": "2026-07-14T11:30:00.000000000Z",
		"ever_succeeded": true
	}`)
	record, err := backupstatus.ParseDiskRecord(data, fixedNow())
	if err != nil {
		t.Fatalf("ParseDiskRecord() error: %v", err)
	}
	got := backupstatus.ProjectDiskRecord(record)
	if got.Status != backupstatus.StatusSucceeded || got.EverSucceeded == nil || !*got.EverSucceeded {
		t.Fatalf("project = %+v", got)
	}
}

func TestParseDiskRecordFailedPreservesEverSucceeded(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"last_outcome": "failed",
		"completed_at": "2026-07-14T11:30:00.000000000Z",
		"ever_succeeded": true,
		"failure_class": "restart"
	}`)
	record, err := backupstatus.ParseDiskRecord(data, fixedNow())
	if err != nil {
		t.Fatalf("ParseDiskRecord() error: %v", err)
	}
	got := backupstatus.ProjectDiskRecord(record)
	if got.Status != backupstatus.StatusFailed || got.Failure == nil || *got.Failure != backupstatus.FailureRestart {
		t.Fatalf("project = %+v", got)
	}
}

func TestParseDiskRecordRejectsInvalidInput(t *testing.T) {
	cases := []string{
		"",
		"{}",
		`{"schema_version":2,"last_outcome":"succeeded","completed_at":"2026-07-14T11:30:00.000000000Z","ever_succeeded":true}`,
		`{"schema_version":1,"last_outcome":"succeeded","completed_at":"2026-07-14T11:30:00.000000000Z","ever_succeeded":false}`,
		`{"schema_version":1,"last_outcome":"failed","completed_at":"2026-07-14T11:30:00.000000000Z","ever_succeeded":false}`,
		`{"schema_version":1,"last_outcome":"failed","completed_at":"2019-01-01T00:00:00Z","ever_succeeded":false,"failure_class":"artifact"}`,
		`{"schema_version":1,"last_outcome":"succeeded","completed_at":"2099-01-01T00:00:00Z","ever_succeeded":true}`,
		`{"schema_version":1,"last_outcome":"succeeded","completed_at":"2026-07-14T11:30:00.000000000Z","ever_succeeded":true,"extra":1}`,
		strings.Repeat("a", 5000),
	}
	for _, input := range cases {
		if _, err := backupstatus.ParseDiskRecord([]byte(input), fixedNow()); err == nil {
			t.Fatalf("ParseDiskRecord(%q) expected error", input)
		}
	}
}

func TestEncodeDiskRecordRespectsWriteLimit(t *testing.T) {
	record := backupstatus.BuildSucceededRecord(fixedNow())
	data, err := backupstatus.EncodeDiskRecord(record)
	if err != nil {
		t.Fatalf("EncodeDiskRecord() error: %v", err)
	}
	if len(data) == 0 || len(data) > backupstatus.MaxWriteSize {
		t.Fatalf("encoded size = %d", len(data))
	}
}

func TestNeverRunAndUnavailableShapes(t *testing.T) {
	never := backupstatus.NeverRun()
	if never.Status != backupstatus.StatusNeverRun || never.EverSucceeded == nil || *never.EverSucceeded {
		t.Fatalf("NeverRun() = %+v", never)
	}
	unavailable := backupstatus.Unavailable()
	if unavailable.Status != backupstatus.StatusUnavailable || unavailable.EverSucceeded != nil {
		t.Fatalf("Unavailable() = %+v", unavailable)
	}
}
