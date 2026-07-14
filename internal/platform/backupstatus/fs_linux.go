//go:build linux

package backupstatus

import (
	"os"
	"syscall"
)

func metadataFromInfo(info os.FileInfo) (FileMetadata, error) {
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return FileMetadata{}, os.ErrInvalid
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return FileMetadata{}, os.ErrInvalid
	}
	return FileMetadata{
		UID:  stat.Uid,
		GID:  stat.Gid,
		Mode: info.Mode().Perm(),
	}, nil
}

// ReadFileMetadata inspects a regular-file status sidecar.
func ReadFileMetadata(path string) (FileMetadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return FileMetadata{}, err
	}
	return metadataFromInfo(info)
}
