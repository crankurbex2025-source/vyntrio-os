//go:build linux

package writemedia

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func listPlatformDevices() ([]Device, error) {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("list /sys/block: %w", err)
	}

	var devices []Device
	for _, entry := range entries {
		name := entry.Name()
		if isVirtualBlockDevice(name) {
			continue
		}
		removable := readSysBlockString(name, "removable") == "1"
		sizeSectors := parseUint64(readSysBlockString(name, "size"))
		model := readSysBlockString(name, "device/model")
		if model == "" {
			model = name
		}
		path := filepath.Join("/dev", name)
		mounted, mounts := deviceMountInfo(path)
		if !removable && !looksLikeUSB(name) {
			continue
		}
		devices = append(devices, Device{
			ID:          path,
			Path:        path,
			Name:        strings.TrimSpace(model),
			SizeBytes:   sizeSectors * 512,
			Removable:   removable,
			BusType:     busTypeForBlock(name),
			Mounted:     mounted,
			MountPoints: mounts,
		})
	}
	return devices, nil
}

func looksLikeUSB(block string) bool {
	link := filepath.Join("/sys/block", block, "device")
	target, err := os.Readlink(link)
	if err != nil {
		return false
	}
	return strings.Contains(target, "usb")
}

func busTypeForBlock(block string) string {
	if looksLikeUSB(block) {
		return "usb"
	}
	return "block"
}

func openDeviceForWrite(devicePath string) (*os.File, error) {
	if err := requireElevatedPrivileges(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open device %s: %w", devicePath, err)
	}
	return f, nil
}

func syncDevice(devicePath string) error {
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("sync device: %w", err)
	}
	defer func() { _ = file.Close() }()
	return file.Sync()
}

func hashDevice(devicePath string, limit uint64) (string, error) {
	if err := requireElevatedPrivileges(); err != nil {
		return "", err
	}
	f, err := os.Open(devicePath)
	if err != nil {
		return "", fmt.Errorf("open device: %w", err)
	}
	defer func() { _ = f.Close() }()

	tmp, err := os.CreateTemp("", "vyntrio-write-media-verify-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := copyLimited(tmp, f, limit); err != nil {
		return "", fmt.Errorf("read device: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	return hashFile(tmpPath)
}

func requireElevatedPrivileges() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("elevated privileges required (re-run with sudo)")
	}
	return nil
}
