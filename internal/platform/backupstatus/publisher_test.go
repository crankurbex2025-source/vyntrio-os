package backupstatus_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

func TestPublisherAtomicSucceededRecord(t *testing.T) {
	dir := t.TempDir()
	completedAt := time.Date(2026, 7, 14, 11, 30, 0, 0, time.UTC)
	publisher := backupstatus.NewPublisher(backupstatus.PublisherDeps{
		GroupGID: func() (uint32, error) { return testGID, nil },
		Chown:    func(string, int, int) error { return nil },
	})
	if err := publisher.PublishSucceeded(context.Background(), dir, completedAt); err != nil {
		t.Fatalf("PublishSucceeded() error: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(dir, "backup-last-run.json.tmp")); !os.IsNotExist(err) {
		t.Fatal("expected temp file removed")
	}
	data, err := os.ReadFile(backupstatus.StatusPath(dir))
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	record, err := backupstatus.ParseDiskRecord(data, completedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("ParseDiskRecord() error: %v", err)
	}
	if record.LastOutcome != backupstatus.OutcomeSucceeded || !record.EverSucceeded {
		t.Fatalf("record = %+v", record)
	}
}

func TestPublisherFailedPreservesPriorSuccess(t *testing.T) {
	dir := t.TempDir()
	firstAt := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	secondAt := time.Date(2026, 7, 14, 11, 0, 0, 0, time.UTC)
	publisher := backupstatus.NewPublisher(backupstatus.PublisherDeps{
		GroupGID: func() (uint32, error) { return testGID, nil },
		Chown:    func(string, int, int) error { return nil },
	})
	if err := publisher.PublishSucceeded(context.Background(), dir, firstAt); err != nil {
		t.Fatal(err)
	}
	if err := publisher.PublishFailed(context.Background(), dir, secondAt, backupstatus.FailureRestart); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(backupstatus.StatusPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	record, err := backupstatus.ParseDiskRecord(data, secondAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if record.LastOutcome != backupstatus.OutcomeFailed || !record.EverSucceeded || record.FailureClass != backupstatus.FailureRestart {
		t.Fatalf("record = %+v", record)
	}
}

func TestPublisherRejectsSymlinkTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")
	if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	final := backupstatus.StatusPath(dir)
	if err := os.Symlink(target, final); err != nil {
		t.Fatal(err)
	}
	publisher := backupstatus.NewPublisher(backupstatus.PublisherDeps{
		GroupGID: func() (uint32, error) { return testGID, nil },
		Chown:    func(string, int, int) error { return nil },
	})
	err := publisher.PublishSucceeded(context.Background(), dir, fixedNow())
	if err == nil {
		t.Fatal("expected symlink rejection")
	}
}
