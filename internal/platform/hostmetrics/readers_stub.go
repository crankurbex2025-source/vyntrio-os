//go:build !linux

package hostmetrics

import "fmt"

type unavailableReader struct{}

func (unavailableReader) ReadLoadAvg() (string, error) {
	return "", fmt.Errorf("loadavg: unsupported platform")
}

func (unavailableReader) ReadMemInfo() (string, error) {
	return "", fmt.Errorf("meminfo: unsupported platform")
}

type unavailableStatFSReader struct{}

func (unavailableStatFSReader) StatStateFilesystem(_ string) (StatFSResult, error) {
	return StatFSResult{}, fmt.Errorf("statfs: unsupported platform")
}

func defaultLoadAvgReader() LoadAvgReader {
	return unavailableReader{}
}

func defaultMemInfoReader() MemInfoReader {
	return unavailableReader{}
}

func defaultStatFSReader() StatFSReader {
	return unavailableStatFSReader{}
}
