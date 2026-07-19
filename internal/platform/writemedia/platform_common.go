package writemedia

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ListDevices returns removable/external block devices suitable for USB install media.
func ListDevices() ([]Device, error) {
	return listPlatformDevices()
}

func readSysBlockString(block, leaf string) string {
	data, err := os.ReadFile(filepath.Join("/sys/block", block, leaf))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func parseUint64(value string) uint64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func isVirtualBlockDevice(name string) bool {
	switch {
	case strings.HasPrefix(name, "loop"):
		return true
	case strings.HasPrefix(name, "ram"):
		return true
	case strings.HasPrefix(name, "dm-"):
		return true
	case strings.HasPrefix(name, "md"):
		return true
	case strings.HasPrefix(name, "nbd"):
		return true
	case strings.HasPrefix(name, "zram"):
		return true
	default:
		return false
	}
}

func deviceMountInfo(devicePath string) (bool, []string) {
	mounts, err := os.Open("/proc/mounts")
	if err != nil {
		return false, nil
	}
	defer func() { _ = mounts.Close() }()

	var points []string
	scanner := bufio.NewScanner(mounts)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		source := fields[0]
		if source == devicePath || strings.HasPrefix(source, devicePath) {
			points = append(points, fields[1])
		}
	}
	return len(points) > 0, points
}

func copyLimited(dst *os.File, src *os.File, limit uint64) (int64, error) {
	buf := make([]byte, 4*1024*1024)
	var written int64
	for written < int64(limit) {
		toRead := int64(len(buf))
		remaining := int64(limit) - written
		if toRead > remaining {
			toRead = remaining
		}
		n, err := src.Read(buf[:toRead])
		if n > 0 {
			wn, werr := dst.Write(buf[:n])
			written += int64(wn)
			if werr != nil {
				return written, werr
			}
			if wn != n {
				return written, fmt.Errorf("short write")
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return written, err
		}
	}
	return written, nil
}
