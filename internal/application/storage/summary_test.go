package storage_test

import (
	"testing"

	appstorage "github.com/crankurbex2025-source/vyntrio-os/internal/application/storage"
)

func TestSummarizeInventory(t *testing.T) {
	size := uint64(1000)
	inventory := appstorage.DisksResponse{
		Status: appstorage.InventoryStatusOK,
		Disks: []appstorage.DiskDevice{
			{ID: "disk-a", Status: "ok", Eligibility: "eligible", SizeBytes: &size},
			{ID: "disk-b", Status: "ok", Eligibility: "excluded", Reasons: []string{"root_disk"}},
			{ID: "disk-c", Status: "ok", Eligibility: "unknown"},
		},
	}

	summary := appstorage.SummarizeInventory(inventory)
	if summary.DiskCount != 3 || summary.EligibleCount != 1 || summary.ExcludedCount != 1 || summary.UnknownCount != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.PoolCount != 0 || summary.ShareCount != 0 || !summary.MutationAvailable {
		t.Fatalf("pools/shares/mutation must stay honest: %+v", summary)
	}
}
