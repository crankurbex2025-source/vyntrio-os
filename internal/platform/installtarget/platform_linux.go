//go:build linux

package installtarget

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type execMountController struct{}

func defaultMountController() MountController {
	return execMountController{}
}

func (execMountController) Mount(devicePath, mountPoint, fsType string) error {
	cmd := exec.Command("mount", "-t", fsType, devicePath, mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount %s: %s", devicePath, strings.TrimSpace(string(output)))
	}
	return nil
}

func (execMountController) Unmount(mountPoint string) error {
	cmd := exec.Command("umount", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("umount %s: %s", mountPoint, strings.TrimSpace(string(output)))
	}
	return nil
}

type blkidProber struct{}

func defaultFSProber() FSProber {
	return blkidProber{}
}

func (blkidProber) ProbeFSType(devicePath string) (string, error) {
	cmd := exec.Command("blkid", "-o", "value", "-s", "TYPE", devicePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	fsType := strings.TrimSpace(string(output))
	switch fsType {
	case "ext3", "ext2":
		return "ext4", nil
	default:
		return fsType, nil
	}
}

type sysfsPartitionLister struct{}

func defaultPartitionLister() PartitionLister {
	return sysfsPartitionLister{}
}

func (sysfsPartitionLister) ListPartitions(kernelName string) ([]string, error) {
	base := filepath.Join("/sys/block", kernelName)
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	var parts []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, kernelName) {
			continue
		}
		partPath := filepath.Join(base, name)
		if _, err := os.Stat(filepath.Join(partPath, "partition")); err != nil {
			continue
		}
		parts = append(parts, name)
	}
	return parts, nil
}
