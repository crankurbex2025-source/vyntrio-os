package hostmetrics

// LoadAvgReader returns the raw contents of /proc/loadavg.
type LoadAvgReader interface {
	ReadLoadAvg() (string, error)
}

// MemInfoReader returns the raw contents of /proc/meminfo.
type MemInfoReader interface {
	ReadMemInfo() (string, error)
}

// StatFSReader returns filesystem capacity for the validated state directory only.
type StatFSReader interface {
	StatStateFilesystem(stateDir string) (StatFSResult, error)
}

// StatFSResult holds statfs-derived capacity and optional mapped filesystem type.
type StatFSResult struct {
	TotalBytes     uint64
	AvailableBytes uint64
	UsedBytes      uint64
	FSType         string
}
