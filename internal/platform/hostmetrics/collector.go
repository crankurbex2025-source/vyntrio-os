package hostmetrics

import (
	"context"
	"runtime"
)

// CollectorDeps configures injectable host metric readers for tests.
type CollectorDeps struct {
	LoadAvg      LoadAvgReader
	MemInfo      MemInfoReader
	StatFS       StatFSReader
	LogicalCores func() int
}

// Collector assembles read-only host metrics for the overview DTO.
type Collector struct {
	stateDir string
	loadAvg  LoadAvgReader
	memInfo  MemInfoReader
	statFS   StatFSReader
	cores    func() int
}

// NewCollector creates a host metrics collector for the validated state directory.
func NewCollector(stateDir string, deps CollectorDeps) Collector {
	loadAvg := deps.LoadAvg
	if loadAvg == nil {
		loadAvg = defaultLoadAvgReader()
	}
	memInfo := deps.MemInfo
	if memInfo == nil {
		memInfo = defaultMemInfoReader()
	}
	statFS := deps.StatFS
	if statFS == nil {
		statFS = defaultStatFSReader()
	}
	cores := deps.LogicalCores
	if cores == nil {
		cores = runtime.NumCPU
	}
	return Collector{
		stateDir: stateDir,
		loadAvg:  loadAvg,
		memInfo:  memInfo,
		statFS:   statFS,
		cores:    cores,
	}
}

// Collect returns host metrics with per-section unavailable degradation.
func (c Collector) Collect(_ context.Context) Host {
	return Host{
		CPU:         c.collectCPU(),
		Memory:      c.collectMemory(),
		Filesystems: []Filesystem{c.collectStateFilesystem()},
	}
}

func (c Collector) collectCPU() CPU {
	cores := c.cores()
	if cores <= 0 {
		return CPU{Status: StatusUnavailable}
	}
	content, err := c.loadAvg.ReadLoadAvg()
	if err != nil {
		return CPU{Status: StatusUnavailable}
	}
	load1m, err := ParseLoadAvg1m(content)
	if err != nil {
		return CPU{Status: StatusUnavailable}
	}
	logicalCores := cores
	load := load1m
	return CPU{
		Status:       StatusOK,
		LogicalCores: &logicalCores,
		Load1m:       &load,
	}
}

func (c Collector) collectMemory() Memory {
	content, err := c.memInfo.ReadMemInfo()
	if err != nil {
		return Memory{Status: StatusUnavailable}
	}
	totalBytes, availableBytes, err := ParseMeminfo(content)
	if err != nil {
		return Memory{Status: StatusUnavailable}
	}
	usedBytes, err := DeriveUsedBytes(totalBytes, availableBytes)
	if err != nil {
		return Memory{Status: StatusUnavailable}
	}
	total := totalBytes
	available := availableBytes
	used := usedBytes
	return Memory{
		Status:         StatusOK,
		TotalBytes:     &total,
		AvailableBytes: &available,
		UsedBytes:      &used,
	}
}

func (c Collector) collectStateFilesystem() Filesystem {
	result, err := c.statFS.StatStateFilesystem(c.stateDir)
	if err != nil {
		return Filesystem{
			ID:     StateFilesystemID,
			Status: StatusUnavailable,
		}
	}
	if result.TotalBytes == 0 {
		return Filesystem{
			ID:     StateFilesystemID,
			Status: StatusUnavailable,
		}
	}
	total := result.TotalBytes
	available := result.AvailableBytes
	used := result.UsedBytes
	fsType := result.FSType
	return Filesystem{
		ID:             StateFilesystemID,
		Status:         StatusOK,
		TotalBytes:     &total,
		AvailableBytes: &available,
		UsedBytes:      &used,
		FSType:         &fsType,
	}
}
