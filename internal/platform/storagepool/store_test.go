package storagepool

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreatePoolAndDatasetAndShare(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	eligible := map[string]bool{"disk-aaa": true, "disk-bbb": true}

	pool, err := store.CreatePool(CreatePoolInput{
		Name:    "tank",
		DiskIDs: []string{"disk-aaa"},
		Confirm: true,
		Now:     time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC),
		NewID:   func() string { return "pool-1" },
	}, eligible)
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}
	if pool.DiskFormatState != DiskFormatPending {
		t.Fatalf("format state = %q", pool.DiskFormatState)
	}
	if pool.Status != StatusDeclared {
		t.Fatalf("status = %q", pool.Status)
	}

	_, ds, err := store.AddDataset(AddDatasetInput{
		PoolID: "pool-1",
		Name:   "media",
		Now:    time.Date(2026, 7, 17, 8, 1, 0, 0, time.UTC),
		NewID:  func() string { return "ds-1" },
	})
	if err != nil {
		t.Fatalf("AddDataset: %v", err)
	}
	if ds.PathIntent != "/tank/media" {
		t.Fatalf("path = %q", ds.PathIntent)
	}

	share, err := store.AddShare(AddShareInput{
		Name:      "media",
		PoolID:    "pool-1",
		DatasetID: "ds-1",
		Protocol:  "smb",
		Now:       time.Date(2026, 7, 17, 8, 2, 0, 0, time.UTC),
		NewID:     func() string { return "share-1" },
	})
	if err != nil {
		t.Fatalf("AddShare: %v", err)
	}
	if share.Protocol != "planned" {
		t.Fatalf("protocol = %q want planned", share.Protocol)
	}

	store2 := NewStore(dir)
	pools, err := store2.ListPools()
	if err != nil || len(pools) != 1 || len(pools[0].Datasets) != 1 {
		t.Fatalf("reload pools = %+v err=%v", pools, err)
	}
	shares, err := store2.ListShares()
	if err != nil || len(shares) != 1 {
		t.Fatalf("reload shares = %+v err=%v", shares, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "storage", storeFileName)); err != nil {
		t.Fatalf("store file: %v", err)
	}
}

func TestCreatePoolRejectsIneligibleAndUnconfirmed(t *testing.T) {
	store := NewStore(t.TempDir())
	eligible := map[string]bool{"disk-aaa": true}
	if _, err := store.CreatePool(CreatePoolInput{Name: "tank", DiskIDs: []string{"disk-aaa"}, Confirm: false}, eligible); err != ErrConfirmRequired {
		t.Fatalf("confirm err = %v", err)
	}
	if _, err := store.CreatePool(CreatePoolInput{Name: "tank", DiskIDs: []string{"disk-zzz"}, Confirm: true}, eligible); err == nil {
		t.Fatal("expected ineligible error")
	}
}
