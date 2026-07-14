package backupstatus

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// FileMetadata captures ownership and mode checks for the status sidecar.
type FileMetadata struct {
	UID  uint32
	GID  uint32
	Mode os.FileMode
}

// GroupResolver resolves the vyntrio group GID.
type GroupResolver func() (uint32, error)

// DefaultGroupResolver looks up the vyntrio group.
func DefaultGroupResolver() GroupResolver {
	return func() (uint32, error) {
		group, err := user.LookupGroup(vyntrioGroupName)
		if err != nil {
			return 0, err
		}
		gid, err := strconv.Atoi(group.Gid)
		if err != nil {
			return 0, err
		}
		return uint32(gid), nil
	}
}

// StatusPath returns the fixed status file path under stateDir.
func StatusPath(stateDir string) string {
	return filepath.Join(stateDir, StatusFileName)
}

// StatusTempPath returns the temporary publication path under stateDir.
func StatusTempPath(stateDir string) string {
	return filepath.Join(stateDir, statusTempName)
}

// ValidateReadableMetadata enforces root:vyntrio 0640 non-writable group/other semantics.
func ValidateReadableMetadata(meta FileMetadata, vyntrioGID uint32) error {
	if meta.UID != 0 {
		return os.ErrPermission
	}
	if meta.GID != vyntrioGID {
		return os.ErrPermission
	}
	if meta.Mode != statusFileMode {
		return os.ErrPermission
	}
	return nil
}
