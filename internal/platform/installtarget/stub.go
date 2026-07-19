package installtarget

import (
	"errors"
	"fmt"
)

// StubMountController records mount operations for tests.
type StubMountController struct {
	MountFn   func(devicePath, mountPoint, fsType string) error
	UnmountFn func(mountPoint string) error
	Mounted   map[string]MountCandidate
}

func (s *StubMountController) Mount(devicePath, mountPoint, fsType string) error {
	if s.MountFn != nil {
		return s.MountFn(devicePath, mountPoint, fsType)
	}
	if s.Mounted == nil {
		s.Mounted = make(map[string]MountCandidate)
	}
	s.Mounted[mountPoint] = MountCandidate{DevicePath: devicePath, FSType: fsType}
	return nil
}

func (s *StubMountController) Unmount(mountPoint string) error {
	if s.UnmountFn != nil {
		return s.UnmountFn(mountPoint)
	}
	delete(s.Mounted, mountPoint)
	return nil
}

// StubFSProber maps device paths to filesystem types in tests.
type StubFSProber struct {
	Types map[string]string
	Err   error
}

func (s StubFSProber) ProbeFSType(devicePath string) (string, error) {
	if s.Err != nil {
		return "", s.Err
	}
	if s.Types == nil {
		return "", errors.New("no filesystem")
	}
	fsType, ok := s.Types[devicePath]
	if !ok || fsType == "" {
		return "", fmt.Errorf("no filesystem on %s", devicePath)
	}
	return fsType, nil
}
