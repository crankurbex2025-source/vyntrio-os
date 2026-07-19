//go:build darwin

package writemedia

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func listPlatformDevices() ([]Device, error) {
	cmd := exec.Command("diskutil", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list: %w", err)
	}
	return parseDiskutilList(out)
}

func parseDiskutilList(data []byte) ([]Device, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var devices []Device
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "/dev/disk") {
			continue
		}
		if !strings.Contains(line, "external") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		devPath := strings.TrimSuffix(fields[0], ":")
		id := strings.TrimPrefix(devPath, "/dev/")
		devices = append(devices, Device{
			ID:        id,
			Path:      rawDiskPath(id),
			Name:      line,
			Removable: true,
			BusType:   "external",
		})
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no external disks found; connect a USB stick and retry")
	}
	return devices, nil
}

func rawDiskPath(id string) string {
	if strings.HasPrefix(id, "/dev/") {
		id = strings.TrimPrefix(id, "/dev/")
	}
	if strings.HasPrefix(id, "rdisk") {
		return "/dev/" + id
	}
	return "/dev/r" + strings.TrimPrefix(id, "disk")
}

func openDeviceForWrite(devicePath string) (*os.File, error) {
	if err := requireElevatedPrivileges(); err != nil {
		return nil, err
	}
	path := devicePath
	if strings.HasPrefix(path, "/dev/disk") && !strings.HasPrefix(path, "/dev/rdisk") {
		path = "/dev/r" + strings.TrimPrefix(path, "/dev/d")
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open device %s: %w", path, err)
	}
	return f, nil
}

func syncDevice(devicePath string) error {
	path := devicePath
	if strings.HasPrefix(path, "/dev/disk") && !strings.HasPrefix(path, "/dev/rdisk") {
		path = "/dev/r" + strings.TrimPrefix(path, "/dev/d")
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0)
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
	path := devicePath
	if strings.HasPrefix(path, "/dev/disk") && !strings.HasPrefix(path, "/dev/rdisk") {
		path = "/dev/r" + strings.TrimPrefix(path, "/dev/d")
	}
	f, err := os.Open(path)
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
