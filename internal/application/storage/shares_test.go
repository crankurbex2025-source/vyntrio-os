package storage_test

import (
	"context"
	"testing"

	appstorage "github.com/crankurbex2025-source/vyntrio-os/internal/application/storage"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storagepool"
)

func TestSharesLoaderReturnsPlannedManagement(t *testing.T) {
	loader := appstorage.NewSharesLoader(stubInventoryLoader{
		response: appstorage.DisksResponse{Status: appstorage.InventoryStatusOK},
	}, storagepool.NewStore(t.TempDir()))
	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.ShareManagement != appstorage.ShareManagementPlanned || !got.MutationAvailable {
		t.Fatalf("share management = %q mutation=%v", got.ShareManagement, got.MutationAvailable)
	}
	if got.ProtocolSupport != "not_available" {
		t.Fatalf("protocol_support = %q", got.ProtocolSupport)
	}
	if len(got.Shares) != 0 {
		t.Fatalf("shares = %v", got.Shares)
	}
}
