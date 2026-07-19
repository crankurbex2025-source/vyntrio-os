package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

func preserveDirName(now time.Time) string {
	return preservePrefix + now.UTC().Format("20060102T150405.000000000Z")
}

func createPreservationCopy(opts Options, now time.Time, members map[string]struct{}) (string, error) {
	preserveRoot := opts.preserveRoot()
	name := preserveDirName(now)
	dir := filepath.Join(preserveRoot, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("%w: create preserve directory", ErrPreservationFailed)
	}
	for member := range members {
		hostPath, ok := opts.hostPathForMember(member)
		if !ok {
			continue
		}
		if err := preserveHostFile(hostPath, filepath.Join(dir, member)); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func preserveHostFile(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("%w: inspect live file", ErrPreservationFailed)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("%w: live file is not a regular file", ErrPreservationFailed)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return fmt.Errorf("%w: create preserve path", ErrPreservationFailed)
	}
	if err := copyFile(src, dst, info.Mode().Perm()); err != nil {
		return fmt.Errorf("%w: copy live file", ErrPreservationFailed)
	}
	return nil
}

func placementTargets(payloads map[string][]byte) map[string]struct{} {
	targets := map[string]struct{}{
		backup.StateDBMember: {},
	}
	for member := range payloads {
		if member == backup.ConfigMember {
			targets[member] = struct{}{}
		}
	}
	for _, sidecar := range []string{backup.StateJournalMem, backup.StateWALMember, backup.StateSHMMember} {
		if _, ok := payloads[sidecar]; ok {
			targets[sidecar] = struct{}{}
		}
	}
	return targets
}

func placePayloads(opts Options, payloads map[string][]byte) (OwnershipTargets, error) {
	targets := OwnershipTargets{StateRoot: opts.StateRoot}
	for member, data := range payloads {
		hostPath, ok := opts.hostPathForMember(member)
		if !ok {
			return OwnershipTargets{}, fmt.Errorf("%w: disallowed member %q", ErrPlacementFailed, member)
		}
		if err := os.MkdirAll(filepath.Dir(hostPath), stateDirMode); err != nil {
			return OwnershipTargets{}, fmt.Errorf("%w: prepare target directory", ErrPlacementFailed)
		}
		mode := stateFileMode
		if member == backup.ConfigMember {
			mode = configFileMode
		}
		if err := writeRegularFile(hostPath, data, mode); err != nil {
			return OwnershipTargets{}, err
		}
		switch member {
		case backup.StateDBMember, backup.StateJournalMem, backup.StateWALMember, backup.StateSHMMember:
			targets.StateFiles = append(targets.StateFiles, hostPath)
		case backup.ConfigMember:
			targets.ConfigFile = hostPath
		}
	}
	if err := removeStaleSidecars(opts, payloads); err != nil {
		return OwnershipTargets{}, err
	}
	return targets, nil
}

func removeStaleSidecars(opts Options, payloads map[string][]byte) error {
	for _, sidecar := range []string{backup.StateJournalMem, backup.StateWALMember, backup.StateSHMMember} {
		if _, inArchive := payloads[sidecar]; inArchive {
			continue
		}
		hostPath, ok := opts.hostPathForMember(sidecar)
		if !ok {
			continue
		}
		if err := os.Remove(hostPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("%w: remove stale sidecar", ErrPlacementFailed)
		}
	}
	return nil
}

func writeRegularFile(path string, data []byte, mode os.FileMode) error {
	if err := os.WriteFile(path, data, mode); err != nil {
		return fmt.Errorf("%w: write target file", ErrPlacementFailed)
	}
	return nil
}

func rollbackFromPreserve(preserveDir string, opts Options, members map[string]struct{}) error {
	for member := range members {
		hostPath, ok := opts.hostPathForMember(member)
		if !ok {
			continue
		}
		preservePath := filepath.Join(preserveDir, member)
		info, err := os.Lstat(preservePath)
		if err != nil {
			if os.IsNotExist(err) {
				_ = os.Remove(hostPath)
				continue
			}
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: preserve file invalid", ErrRollbackFailed)
		}
		if err := copyFile(preservePath, hostPath, info.Mode().Perm()); err != nil {
			return fmt.Errorf("%w: restore preserve file", ErrRollbackFailed)
		}
	}
	return nil
}
