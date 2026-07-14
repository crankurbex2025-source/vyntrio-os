package backup_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

func TestSelectSourceMembersRejectsFIFO(t *testing.T) {
	root := t.TempDir()
	stateRoot := filepath.Join(root, "state")
	configPath := filepath.Join(root, "config.toml")
	if err := os.MkdirAll(stateRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateRoot, backup.MainDBName), []byte("db"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("cfg"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateRoot, backup.JournalSidecar), []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(stateRoot, backup.JournalSidecar)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateRoot, backup.JournalSidecar), []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	// Replace journal with a directory to trigger non-regular rejection.
	_ = os.Remove(filepath.Join(stateRoot, backup.JournalSidecar))
	if err := os.Mkdir(filepath.Join(stateRoot, backup.JournalSidecar), 0o750); err != nil {
		t.Fatal(err)
	}
	_, err := backup.SelectSourceMembers(stateRoot, configPath)
	if !errors.Is(err, backup.ErrSourceInvalid) {
		t.Fatalf("err = %v", err)
	}
}
