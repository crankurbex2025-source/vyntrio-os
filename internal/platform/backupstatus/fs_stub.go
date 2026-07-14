//go:build !linux

package backupstatus

import "os"

func metadataFromInfo(os.FileInfo) (FileMetadata, error) {
	return FileMetadata{}, os.ErrInvalid
}

// ReadFileMetadata inspects a regular-file status sidecar.
func ReadFileMetadata(string) (FileMetadata, error) {
	return FileMetadata{}, os.ErrInvalid
}
