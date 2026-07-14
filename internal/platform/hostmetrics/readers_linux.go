//go:build linux

package hostmetrics

import (
	"fmt"
	"os"
)

const maxProcReadBytes = 64 * 1024

type osLoadAvgReader struct{}

func (osLoadAvgReader) ReadLoadAvg() (string, error) {
	return readProcFileBounded("/proc/loadavg", maxLoadAvgInputBytes)
}

type osMemInfoReader struct{}

func (osMemInfoReader) ReadMemInfo() (string, error) {
	return readProcFileBounded("/proc/meminfo", maxMemInfoInputBytes)
}

type osStatFSReader struct{}

func (osStatFSReader) StatStateFilesystem(stateDir string) (StatFSResult, error) {
	if stateDir == "" {
		return StatFSResult{}, fmt.Errorf("statfs: missing state directory")
	}
	return statfsCapacity(stateDir)
}

func readProcFileBounded(path string, maxBytes int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) == 0 || len(data) > maxBytes {
		return "", fmt.Errorf("proc read: invalid size")
	}
	return string(data), nil
}

func defaultLoadAvgReader() LoadAvgReader {
	return osLoadAvgReader{}
}

func defaultMemInfoReader() MemInfoReader {
	return osMemInfoReader{}
}

func defaultStatFSReader() StatFSReader {
	return osStatFSReader{}
}
