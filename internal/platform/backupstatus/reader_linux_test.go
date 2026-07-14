//go:build linux

package backupstatus_test

import (
	"context"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

func TestReaderValidFileReturnsSucceeded(t *testing.T) {
	content := `{
		"schema_version": 1,
		"last_outcome": "succeeded",
		"completed_at": "2026-07-14T11:30:00.000000000Z",
		"ever_succeeded": true
	}`
	reader := backupstatus.NewReader("/state", backupstatus.ReaderDeps{
		Lstat: func(string) (os.FileInfo, error) {
			return linuxFakeFileInfo{size: int64(len(content)), mode: 0o100640, uid: 0, gid: testGID}, nil
		},
		Open: func(string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(content)), nil
		},
		GroupGID: func() (uint32, error) { return testGID, nil },
		Now:      func() time.Time { return fixedNow() },
	})
	got := reader.Read(context.Background())
	if got.Status != backupstatus.StatusSucceeded {
		t.Fatalf("Read() = %+v", got)
	}
}

func TestReaderRejectsUnsafeMetadata(t *testing.T) {
	content := `{
		"schema_version": 1,
		"last_outcome": "succeeded",
		"completed_at": "2026-07-14T11:30:00.000000000Z",
		"ever_succeeded": true
	}`
	reader := backupstatus.NewReader("/state", backupstatus.ReaderDeps{
		Lstat: func(string) (os.FileInfo, error) {
			return linuxFakeFileInfo{size: int64(len(content)), mode: 0o100644, uid: 0, gid: testGID}, nil
		},
		Open: func(string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(content)), nil
		},
		GroupGID: func() (uint32, error) { return testGID, nil },
	})
	got := reader.Read(context.Background())
	if got.Status != backupstatus.StatusUnavailable {
		t.Fatalf("Read() = %+v, want unavailable", got)
	}
}

func TestReaderRejectsOversizedFile(t *testing.T) {
	reader := backupstatus.NewReader("/state", backupstatus.ReaderDeps{
		Lstat: func(string) (os.FileInfo, error) {
			return linuxFakeFileInfo{size: backupstatus.MaxReadSize + 1, mode: 0o100640, uid: 0, gid: testGID}, nil
		},
		Open: func(string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(strings.Repeat("a", backupstatus.MaxReadSize+1))), nil
		},
		GroupGID: func() (uint32, error) { return testGID, nil },
	})
	got := reader.Read(context.Background())
	if got.Status != backupstatus.StatusUnavailable {
		t.Fatalf("Read() = %+v", got)
	}
}

type linuxFakeFileInfo struct {
	size int64
	mode os.FileMode
	uid  uint32
	gid  uint32
}

func (f linuxFakeFileInfo) Name() string       { return backupstatus.StatusFileName }
func (f linuxFakeFileInfo) Size() int64        { return f.size }
func (f linuxFakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f linuxFakeFileInfo) IsDir() bool        { return false }
func (f linuxFakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f linuxFakeFileInfo) Sys() any {
	return &syscall.Stat_t{Uid: f.uid, Gid: f.gid}
}
