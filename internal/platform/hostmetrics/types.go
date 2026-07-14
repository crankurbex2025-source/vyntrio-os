package hostmetrics

const (
	StatusOK          = "ok"
	StatusUnavailable = "unavailable"
	StateFilesystemID = "state"
)

// Host is the safe host metrics section for GET /api/v1/overview.
type Host struct {
	CPU         CPU          `json:"cpu"`
	Memory      Memory       `json:"memory"`
	Filesystems []Filesystem `json:"filesystems"`
}

// CPU reports logical core count and 1-minute load average only.
type CPU struct {
	Status       string   `json:"status"`
	LogicalCores *int     `json:"logical_cores,omitempty"`
	Load1m       *float64 `json:"load_1m,omitempty"`
}

// Memory reports total, available, and derived used bytes.
type Memory struct {
	Status         string  `json:"status"`
	TotalBytes     *uint64 `json:"total_bytes,omitempty"`
	AvailableBytes *uint64 `json:"available_bytes,omitempty"`
	UsedBytes      *uint64 `json:"used_bytes,omitempty"`
}

// Filesystem reports capacity for a fixed allowlisted appliance mount ID.
type Filesystem struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	TotalBytes     *uint64 `json:"total_bytes,omitempty"`
	AvailableBytes *uint64 `json:"available_bytes,omitempty"`
	UsedBytes      *uint64 `json:"used_bytes,omitempty"`
	FSType         *string `json:"fs_type,omitempty"`
}
