package backupstatus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PublisherDeps configures injectable filesystem behavior for tests.
type PublisherDeps struct {
	Lstat    func(string) (os.FileInfo, error)
	OpenFile func(string, int, os.FileMode) (*os.File, error)
	Rename   func(string, string) error
	Chmod    func(string, os.FileMode) error
	Chown    func(string, int, int) error
	Fsync    func(*os.File) error
	FsyncDir func(string) error
	Remove   func(string) error
	ReadFile func(string) ([]byte, error)
	GroupGID GroupResolver
	Now      func() time.Time
}

// Publisher atomically writes the backup status sidecar as root.
type Publisher struct {
	groupGID GroupResolver
	lstat    func(string) (os.FileInfo, error)
	openFile func(string, int, os.FileMode) (*os.File, error)
	rename   func(string, string) error
	chmod    func(string, os.FileMode) error
	chown    func(string, int, int) error
	fsync    func(*os.File) error
	fsyncDir func(string) error
	remove   func(string) error
	readFile func(string) ([]byte, error)
	now      func() time.Time
}

// NewPublisher creates a root-only backup status publisher.
func NewPublisher(deps PublisherDeps) Publisher {
	groupGID := deps.GroupGID
	if groupGID == nil {
		groupGID = DefaultGroupResolver()
	}
	lstat := deps.Lstat
	if lstat == nil {
		lstat = os.Lstat
	}
	openFile := deps.OpenFile
	if openFile == nil {
		openFile = os.OpenFile
	}
	rename := deps.Rename
	if rename == nil {
		rename = os.Rename
	}
	chmod := deps.Chmod
	if chmod == nil {
		chmod = os.Chmod
	}
	chown := deps.Chown
	if chown == nil {
		chown = func(path string, uid, gid int) error {
			return os.Chown(path, uid, gid)
		}
	}
	fsync := deps.Fsync
	if fsync == nil {
		fsync = func(f *os.File) error {
			return f.Sync()
		}
	}
	fsyncDir := deps.FsyncDir
	if fsyncDir == nil {
		fsyncDir = fsyncDirectory
	}
	remove := deps.Remove
	if remove == nil {
		remove = os.Remove
	}
	readFile := deps.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	return Publisher{
		groupGID: groupGID,
		lstat:    lstat,
		openFile: openFile,
		rename:   rename,
		chmod:    chmod,
		chown:    chown,
		fsync:    fsync,
		fsyncDir: fsyncDir,
		remove:   remove,
		readFile: readFile,
		now:      now,
	}
}

// PublishSucceeded writes a succeeded status record.
func (p Publisher) PublishSucceeded(_ context.Context, stateDir string, completedAt time.Time) error {
	record := BuildSucceededRecord(completedAt)
	return p.publish(stateDir, record)
}

// PublishFailed writes a failed status record preserving prior ever_succeeded.
func (p Publisher) PublishFailed(_ context.Context, stateDir string, completedAt time.Time, failureClass string) error {
	if _, ok := allowedFailureClasses[failureClass]; !ok {
		return fmt.Errorf("backup status: invalid failure class")
	}
	prior := p.loadPriorEverSucceeded(stateDir)
	record := BuildFailedRecord(completedAt, failureClass, prior)
	return p.publish(stateDir, record)
}

func (p Publisher) loadPriorEverSucceeded(stateDir string) bool {
	path := StatusPath(stateDir)
	data, err := p.readFile(path)
	if err != nil {
		return false
	}
	return PriorEverSucceeded(data, p.now())
}

func (p Publisher) publish(stateDir string, record DiskRecord) error {
	if err := validateDiskRecord(record, p.now()); err != nil {
		return err
	}
	data, err := EncodeDiskRecord(record)
	if err != nil {
		return err
	}
	if err := p.validateTarget(stateDir); err != nil {
		return err
	}

	vyntrioGID, err := p.groupGID()
	if err != nil {
		return err
	}

	finalPath := StatusPath(stateDir)
	tempPath := StatusTempPath(stateDir)
	_ = p.remove(tempPath)

	file, err := p.openFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, statusFileMode)
	if err != nil {
		return err
	}

	wrote := false
	defer func() {
		_ = file.Close()
		if !wrote {
			_ = p.remove(tempPath)
		}
	}()

	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := p.fsync(file); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	wrote = true

	if err := p.chown(tempPath, 0, int(vyntrioGID)); err != nil {
		return err
	}
	if err := p.chmod(tempPath, statusFileMode); err != nil {
		return err
	}
	if err := p.rename(tempPath, finalPath); err != nil {
		return err
	}
	if err := p.fsyncDir(stateDir); err != nil {
		return err
	}
	return nil
}

func (p Publisher) validateTarget(stateDir string) error {
	finalPath := StatusPath(stateDir)
	info, err := p.lstat(finalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return os.ErrInvalid
	}
	return nil
}

func fsyncDirectory(dir string) error {
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return file.Sync()
}

// EnsureStateDirExists verifies the state directory is usable for publication.
func EnsureStateDirExists(stateDir string) error {
	info, err := os.Lstat(stateDir)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return os.ErrInvalid
	}
	abs, err := filepath.Abs(stateDir)
	if err != nil {
		return err
	}
	if abs != filepath.Clean(abs) {
		return os.ErrInvalid
	}
	return nil
}
