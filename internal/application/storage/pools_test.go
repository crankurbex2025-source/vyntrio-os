package storage_test

import (
	"context"
	"testing"

	appstorage "github.com/crankurbex2025-source/vyntrio-os/internal/application/storage"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storagepool"
)

type stubInventoryLoader struct {
	response appstorage.DisksResponse
	err      error
}

func (s stubInventoryLoader) Load(context.Context) (appstorage.DisksResponse, error) {
	return s.response, s.err
}

func TestPoolsLoaderReturnsDeclaredManagement(t *testing.T) {
	dir := t.TempDir()
	store := storagepool.NewStore(dir)
	loader := appstorage.NewPoolsLoader(stubInventoryLoader{
		response: appstorage.DisksResponse{Status: appstorage.InventoryStatusOK},
	}, store)
	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Status != appstorage.PoolsStatusOK {
		t.Fatalf("status = %q", got.Status)
	}
	if len(got.Pools) != 0 {
		t.Fatalf("pools = %v, want empty", got.Pools)
	}
	if got.PoolManagement != appstorage.PoolManagementDeclared || !got.MutationAvailable || got.DiskFormatApplied {
		t.Fatalf("pool_management/mutation/format = %q %v %v", got.PoolManagement, got.MutationAvailable, got.DiskFormatApplied)
	}
}

func TestCreatePoolServiceDeclaresPool(t *testing.T) {
	dir := t.TempDir()
	store := storagepool.NewStore(dir)
	size := uint64(1000)
	inventory := stubInventoryLoader{
		response: appstorage.DisksResponse{
			Status: appstorage.InventoryStatusOK,
			Disks: []appstorage.DiskDevice{
				{ID: "disk-aaa", Status: "ok", SizeBytes: &size, Eligibility: "eligible"},
				{ID: "disk-root", Status: "ok", Eligibility: "excluded", Reasons: []string{"root_disk"}},
			},
		},
	}
	svc := appstorage.NewCreatePoolService(inventory, store)
	pool, err := svc.Create(context.Background(), appstorage.CreatePoolRequest{
		Name: "tank", DiskIDs: []string{"disk-aaa"}, Confirm: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pool.Name != "tank" || pool.DiskFormatState != storagepool.DiskFormatPending {
		t.Fatalf("pool = %+v", pool)
	}
	loader := appstorage.NewPoolsLoader(inventory, store)
	got, err := loader.Load(context.Background())
	if err != nil || len(got.Pools) != 1 {
		t.Fatalf("Load after create: %+v err=%v", got, err)
	}
}

func TestPoolsLoaderUnavailableWhenInventoryUnavailable(t *testing.T) {
	loader := appstorage.NewPoolsLoader(stubInventoryLoader{
		response: appstorage.DisksResponse{Status: appstorage.InventoryStatusUnavailable},
	}, nil)
	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Status != appstorage.PoolsStatusUnavailable {
		t.Fatalf("status = %q", got.Status)
	}
}
