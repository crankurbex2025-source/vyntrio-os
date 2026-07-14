package hostmetrics

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const maxLoadAvgInputBytes = 256

// ParseLoadAvg1m extracts the 1-minute load average from /proc/loadavg content.
func ParseLoadAvg1m(content string) (float64, error) {
	if len(content) == 0 || len(content) > maxLoadAvgInputBytes {
		return 0, fmt.Errorf("loadavg: invalid input size")
	}
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) < 1 {
		return 0, fmt.Errorf("loadavg: missing field")
	}
	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || math.IsNaN(load) || math.IsInf(load, 0) || load < 0 {
		return 0, fmt.Errorf("loadavg: invalid 1m value")
	}
	return load, nil
}
