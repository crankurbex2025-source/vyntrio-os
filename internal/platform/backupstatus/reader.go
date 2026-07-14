package backupstatus

import (
	"context"
	"errors"
	"io"
	"os"
	"time"
)

// ReaderDeps configures injectable filesystem behavior for tests.
type ReaderDeps struct {
	Lstat    func(string) (os.FileInfo, error)
	Open     func(string) (io.ReadCloser, error)
	GroupGID GroupResolver
	Now      func() time.Time
}

// Reader loads and projects the backup status sidecar for the overview API.
type Reader struct {
	stateDir string
	lstat    func(string) (os.FileInfo, error)
	open     func(string) (io.ReadCloser, error)
	groupGID GroupResolver
	now      func() time.Time
}

// NewReader creates a backup status reader for the validated state directory.
func NewReader(stateDir string, deps ReaderDeps) Reader {
	lstat := deps.Lstat
	if lstat == nil {
		lstat = os.Lstat
	}
	open := deps.Open
	if open == nil {
		open = func(path string) (io.ReadCloser, error) {
			return os.Open(path)
		}
	}
	groupGID := deps.GroupGID
	if groupGID == nil {
		groupGID = DefaultGroupResolver()
	}
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	return Reader{
		stateDir: stateDir,
		lstat:    lstat,
		open:     open,
		groupGID: groupGID,
		now:      now,
	}
}

// Read returns the safe API backup section.
func (r Reader) Read(_ context.Context) Backup {
	path := StatusPath(r.stateDir)
	info, err := r.lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NeverRun()
		}
		return Unavailable()
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Unavailable()
	}
	if info.Size() == 0 || info.Size() > MaxReadSize {
		return Unavailable()
	}

	vyntrioGID, err := r.groupGID()
	if err != nil {
		return Unavailable()
	}
	meta, err := metadataFromInfo(info)
	if err != nil {
		return Unavailable()
	}
	if err := ValidateReadableMetadata(meta, vyntrioGID); err != nil {
		return Unavailable()
	}

	file, err := r.open(path)
	if err != nil {
		return Unavailable()
	}
	defer func() { _ = file.Close() }()

	data, err := readBounded(file, MaxReadSize)
	if err != nil {
		return Unavailable()
	}

	record, err := ParseDiskRecord(data, r.now())
	if err != nil {
		return Unavailable()
	}
	return ProjectDiskRecord(record)
}

func readBounded(r io.Reader, max int) ([]byte, error) {
	limited := io.LimitReader(r, int64(max)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > max {
		return nil, os.ErrInvalid
	}
	return data, nil
}
