//go:build linux

package storageinventory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type linuxBlockDeviceReader struct{}

func platformBlockDeviceReader() BlockDeviceReader {
	return linuxBlockDeviceReader{}
}

func (linuxBlockDeviceReader) ListBlockDevices(stateDir string) ([]RawDevice, error) {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, err
	}

	rootDisk, stateDisk, mounts, err := parseMountContext(stateDir)
	if err != nil {
		return nil, err
	}

	devices := make([]RawDevice, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if shouldSkipBlockDevice(name) {
			continue
		}
		raw, err := readBlockDevice(name, rootDisk, stateDisk, mounts[name])
		if err != nil {
			devices = append(devices, RawDevice{KernelName: name, IdentityAmbiguous: true})
			continue
		}
		devices = append(devices, raw)
	}
	return devices, nil
}

func shouldSkipBlockDevice(name string) bool {
	switch {
	case strings.HasPrefix(name, "loop"):
		return true
	case strings.HasPrefix(name, "ram"):
		return true
	case strings.HasPrefix(name, "dm-"):
		return true
	case strings.HasPrefix(name, "zram"):
		return true
	case strings.HasPrefix(name, "fd"):
		return true
	default:
		return false
	}
}

func readBlockDevice(name, rootDisk, stateDisk string, mountInfo deviceMountInfo) (RawDevice, error) {
	base := filepath.Join("/sys/block", name)
	sizeBytes, sizeKnown, err := readSysfsSize(filepath.Join(base, "size"))
	if err != nil {
		return RawDevice{}, err
	}
	removable, _ := readSysfsBool(filepath.Join(base, "removable"))
	readOnly, _ := readSysfsBool(filepath.Join(base, "ro"))
	rotational, rotKnown := readOptionalSysfsBool(filepath.Join(base, "queue/rotational"))
	optical := strings.HasPrefix(name, "sr")

	raw := RawDevice{
		KernelName: name,
		SizeBytes:  sizeBytes,
		SizeKnown:  sizeKnown,
		Removable:  removable,
		ReadOnly:   readOnly,
		Optical:    optical,
		Mounted:    mountInfo.Mounted,
		FSTypes:    mountInfo.FSTypes,
		RootDisk:   name == rootDisk,
		StateDisk:  name == stateDisk,
	}
	if rotKnown {
		raw.Rotational = &rotational
	}
	return raw, nil
}

type deviceMountInfo struct {
	Mounted bool
	FSTypes []string
}

func parseMountContext(stateDir string) (rootDisk, stateDisk string, mounts map[string]deviceMountInfo, err error) {
	mounts = make(map[string]deviceMountInfo)
	data, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return "", "", nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		mountPoint := unescapeMountinfo(fields[4])
		fsType := ""
		for i := 8; i < len(fields); i++ {
			if fields[i] == "-" && i+1 < len(fields) {
				fsType = fields[i+1]
				break
			}
		}
		disk, err := diskFromMajorMinor(fields[2])
		if err != nil || disk == "" {
			continue
		}
		info := mounts[disk]
		info.Mounted = true
		if fsType != "" {
			info.FSTypes = appendUniqueString(info.FSTypes, normalizeFSType(fsType))
		}
		mounts[disk] = info
		if mountPoint == "/" {
			rootDisk = disk
		}
		if stateDir != "" && mountPoint == stateDir {
			stateDisk = disk
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", nil, err
	}
	return rootDisk, stateDisk, mounts, nil
}

func diskFromMajorMinor(majorMinor string) (string, error) {
	link := filepath.Join("/sys/dev/block", majorMinor)
	target, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	base := filepath.Base(filepath.Clean(target))
	return parentDiskName(base), nil
}

func parentDiskName(name string) string {
	if strings.HasPrefix(name, "nvme") && strings.Contains(name, "p") {
		if idx := strings.LastIndex(name, "p"); idx > 4 {
			return name[:idx]
		}
	}
	for len(name) > 0 && name[len(name)-1] >= '0' && name[len(name)-1] <= '9' {
		name = name[:len(name)-1]
	}
	return name
}

func unescapeMountinfo(value string) string {
	replacer := strings.NewReplacer("\\040", " ", "\\011", "\t")
	return replacer.Replace(value)
}

func normalizeFSType(fsType string) string {
	switch fsType {
	case "ext3", "ext2":
		return "ext4"
	default:
		return fsType
	}
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func readSysfsSize(path string) (uint64, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false, err
	}
	sectors, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, false, err
	}
	return sectors * 512, true, nil
}

func readSysfsBool(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	switch strings.TrimSpace(string(data)) {
	case "1":
		return true, nil
	case "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool in %s", path)
	}
}

func readOptionalSysfsBool(path string) (bool, bool) {
	value, err := readSysfsBool(path)
	if err != nil {
		return false, false
	}
	return value, true
}
