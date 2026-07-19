package installtarget

import (
	"fmt"
	"os"
	"path/filepath"
)

// RollbackUnmount attempts to undo a mounted target session after a failed apply.
func RollbackUnmount(session MountSession, outcome *ApplyOutcome) error {
	if session.mounter == nil || session.MountPoint == "" {
		return nil
	}
	if err := session.mounter.Unmount(session.MountPoint); err != nil {
		if outcome != nil {
			outcome.FailureStage = "rollback"
			outcome.FailureReason = fmt.Sprintf("%s: %v", ReasonRollbackFailed, err)
		}
		return fmt.Errorf("%w: %v", ErrRollbackFailed, err)
	}
	_ = os.RemoveAll(session.MountPoint)
	if outcome != nil {
		outcome.Mounted = false
		outcome.RolledBack = true
	}
	return nil
}

// RollbackCopiedFiles removes payload files written during a partial apply.
func RollbackCopiedFiles(mountPoint string, payloadPaths []string) {
	for _, rel := range payloadPaths {
		_ = os.Remove(filepath.Join(mountPoint, filepath.FromSlash(rel)))
	}
	record := filepath.Join(mountPoint, "INSTALL_RECORD.txt")
	_ = os.Remove(record)
}
