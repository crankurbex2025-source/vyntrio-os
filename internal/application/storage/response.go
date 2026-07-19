package storage

const (
	InventoryStatusOK          = "ok"
	InventoryStatusUnavailable = "unavailable"
)

// DisksResponse is the safe read-only storage inventory for GET /api/v1/storage/disks.
type DisksResponse struct {
	CollectedAt string       `json:"collected_at"`
	Status      string       `json:"status"`
	Disks       []DiskDevice `json:"disks"`
}

// DiskDevice is one classified block device candidate.
type DiskDevice struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	SizeBytes   *uint64  `json:"size_bytes,omitempty"`
	Rotational  *bool    `json:"rotational,omitempty"`
	Removable   *bool    `json:"removable,omitempty"`
	Eligibility string   `json:"eligibility"`
	Reasons     []string `json:"reasons,omitempty"`
}
