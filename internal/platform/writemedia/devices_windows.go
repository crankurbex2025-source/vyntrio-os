//go:build windows

package writemedia

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type psDisk struct {
	Number       int    `json:"Number"`
	FriendlyName string `json:"FriendlyName"`
	Size         uint64 `json:"Size"`
	BusType      string `json:"BusType"`
}

func listPlatformDevices() ([]Device, error) {
	script := `$disks = Get-Disk | Where-Object { $_.BusType -in @('USB','SD') -and $_.OperationalStatus -eq 'Online' }; if ($null -eq $disks) { '[]' } elseif ($disks -is [array]) { $disks | Select-Object Number,FriendlyName,Size,BusType | ConvertTo-Json -Compress } else { @($disks) | Select-Object Number,FriendlyName,Size,BusType | ConvertTo-Json -Compress }`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list disks via PowerShell: %w", err)
	}
	payload := strings.TrimSpace(string(out))
	if payload == "" || payload == "[]" {
		return nil, fmt.Errorf("no USB/SD disks found; connect a USB stick and retry")
	}

	var disks []psDisk
	if payload[0] == '{' {
		var one psDisk
		if err := json.Unmarshal([]byte(payload), &one); err != nil {
			return nil, fmt.Errorf("decode disk list: %w", err)
		}
		disks = []psDisk{one}
	} else {
		if err := json.Unmarshal([]byte(payload), &disks); err != nil {
			return nil, fmt.Errorf("decode disk list: %w", err)
		}
	}

	devices := make([]Device, 0, len(disks))
	for _, disk := range disks {
		path := physicalDrivePath(disk.Number)
		devices = append(devices, Device{
			ID:        path,
			Path:      path,
			Name:      strings.TrimSpace(disk.FriendlyName),
			SizeBytes: disk.Size,
			Removable: true,
			BusType:   strings.ToLower(strings.TrimSpace(disk.BusType)),
		})
	}
	return devices, nil
}

func physicalDrivePath(number int) string {
	return `\\.\PhysicalDrive` + strconv.Itoa(number)
}

func openDeviceForWrite(devicePath string) (*os.File, error) {
	if err := requireElevatedPrivileges(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open device %s: %w (run as Administrator)", devicePath, err)
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
	// Best-effort: actual write will fail without admin; avoid cgo-only APIs here.
	return nil
}
