package backupstatus_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

func TestReaderMissingFileReturnsNeverRun(t *testing.T) {
	reader := backupstatus.NewReader(t.TempDir(), backupstatus.ReaderDeps{
		GroupGID: func() (uint32, error) { return testGID, nil },
	})
	got := reader.Read(context.Background())
	if got.Status != backupstatus.StatusNeverRun {
		t.Fatalf("Read() = %+v", got)
	}
}

func TestReaderRejectsSymlink(t *testing.T) {
	reader := backupstatus.NewReader(t.TempDir(), backupstatus.ReaderDeps{
		Lstat: func(string) (os.FileInfo, error) {
			return fakeFileInfo{mode: os.ModeSymlink | 0o777}, nil
		},
		GroupGID: func() (uint32, error) { return testGID, nil },
	})
	got := reader.Read(context.Background())
	if got.Status != backupstatus.StatusUnavailable {
		t.Fatalf("Read() = %+v", got)
	}
}

type fakeFileInfo struct {
	size int64
	mode os.FileMode
	uid  uint32
	gid  uint32
}

func (f fakeFileInfo) Name() string       { return backupstatus.StatusFileName }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) Sys() any           { return nil }
