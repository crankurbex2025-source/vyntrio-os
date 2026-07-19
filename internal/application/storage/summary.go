package storage

const (
	SummaryStatusOK          = "ok"
	SummaryStatusUnavailable = "unavailable"
)

// Summary is a read-only aggregate of block inventory for overview and storage layout.
type Summary struct {
	Status            string `json:"status"`
	DiskCount         int    `json:"disk_count"`
	EligibleCount     int    `json:"eligible_count"`
	ExcludedCount     int    `json:"excluded_count"`
	UnknownCount      int    `json:"unknown_count"`
	PoolCount         int    `json:"pool_count"`
	ShareCount        int    `json:"share_count"`
	MutationAvailable bool   `json:"mutation_available"`
}

// SummarizeInventory derives overview-safe storage counts from a disk inventory response.
// Pool/share counts remain 0 here; use SummarizeLayout when declared pools are loaded.
func SummarizeInventory(inventory DisksResponse) Summary {
	summary := Summary{
		Status:            inventory.Status,
		MutationAvailable: true,
		PoolCount:         0,
		ShareCount:        0,
	}
	if inventory.Status != InventoryStatusOK {
		summary.Status = SummaryStatusUnavailable
		return summary
	}
	summary.Status = SummaryStatusOK
	summary.DiskCount = len(inventory.Disks)
	for _, disk := range inventory.Disks {
		switch disk.Eligibility {
		case "eligible":
			summary.EligibleCount++
		case "excluded":
			summary.ExcludedCount++
		case "unknown":
			summary.UnknownCount++
		}
	}
	return summary
}

// SummarizeLayout merges inventory counts with declared pool/share counts.
func SummarizeLayout(inventory DisksResponse, poolCount, shareCount int) Summary {
	summary := SummarizeInventory(inventory)
	summary.PoolCount = poolCount
	summary.ShareCount = shareCount
	summary.MutationAvailable = true
	return summary
}
