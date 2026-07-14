package hostmetrics

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	maxMemInfoInputBytes = 64 * 1024
	maxMemInfoLines      = 512
	kibToBytes           = 1024
)

// ParseMeminfo extracts MemTotal and MemAvailable in bytes from /proc/meminfo content.
func ParseMeminfo(content string) (totalBytes, availableBytes uint64, err error) {
	if len(content) == 0 || len(content) > maxMemInfoInputBytes {
		return 0, 0, fmt.Errorf("meminfo: invalid input size")
	}

	var memTotalKib, memAvailableKib uint64
	var foundTotal, foundAvailable bool
	lines := 0

	for line := range strings.SplitSeq(content, "\n") {
		lines++
		if lines > maxMemInfoLines {
			return 0, 0, fmt.Errorf("meminfo: too many lines")
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			memTotalKib, err = parseMeminfoKibValue(line, "MemTotal:")
			if err != nil {
				return 0, 0, err
			}
			foundTotal = true
		case strings.HasPrefix(line, "MemAvailable:"):
			memAvailableKib, err = parseMeminfoKibValue(line, "MemAvailable:")
			if err != nil {
				return 0, 0, err
			}
			foundAvailable = true
		}
		if foundTotal && foundAvailable {
			break
		}
	}

	if !foundTotal || !foundAvailable {
		return 0, 0, fmt.Errorf("meminfo: required fields missing")
	}
	if memTotalKib == 0 {
		return 0, 0, fmt.Errorf("meminfo: invalid total")
	}
	if memAvailableKib > memTotalKib {
		return 0, 0, fmt.Errorf("meminfo: available exceeds total")
	}

	totalBytes, err = kibToBytesSafe(memTotalKib)
	if err != nil {
		return 0, 0, err
	}
	availableBytes, err = kibToBytesSafe(memAvailableKib)
	if err != nil {
		return 0, 0, err
	}
	return totalBytes, availableBytes, nil
}

func parseMeminfoKibValue(line, prefix string) (uint64, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if rest == "" {
		return 0, fmt.Errorf("meminfo: empty value")
	}
	fields := strings.Fields(rest)
	if len(fields) < 2 || fields[1] != "kB" {
		return 0, fmt.Errorf("meminfo: invalid unit")
	}
	value, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("meminfo: invalid value")
	}
	return value, nil
}

func kibToBytesSafe(kib uint64) (uint64, error) {
	if kib > mathMaxUint64/kibToBytes {
		return 0, fmt.Errorf("meminfo: overflow")
	}
	return kib * kibToBytes, nil
}

const mathMaxUint64 = ^uint64(0)
