package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storagepool"
)

const (
	PoolsStatusOK              = "ok"
	PoolsStatusUnavailable     = "unavailable"
	PoolManagementDeclared     = "declared_pools"
	PoolManagementNotAvailable = "not_available"
)

// PoolView is one pool in the GET /api/v1/storage/pools response.
type PoolView struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Status          string         `json:"status"`
	DiskIDs         []string       `json:"disk_ids"`
	DiskFormatState string         `json:"disk_format_state"`
	Datasets        []DatasetView  `json:"datasets"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}

// DatasetView is a planned dataset under a pool.
type DatasetView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PathIntent string `json:"path_intent"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// PoolsResponse is the storage pools view for GET /api/v1/storage/pools.
type PoolsResponse struct {
	CollectedAt       string     `json:"collected_at"`
	Status            string     `json:"status"`
	InventoryStatus   string     `json:"inventory_status"`
	Pools             []PoolView `json:"pools"`
	PoolManagement    string     `json:"pool_management"`
	MutationAvailable bool       `json:"mutation_available"`
	DiskFormatApplied bool       `json:"disk_format_applied"`
	Note              string     `json:"note,omitempty"`
}

// PoolStore reads and writes declared pools.
type PoolStore interface {
	ListPools() ([]storagepool.Pool, error)
	ListShares() ([]storagepool.Share, error)
	UsedDiskIDs() (map[string]string, error)
	CreatePool(input storagepool.CreatePoolInput, eligible map[string]bool) (storagepool.Pool, error)
	AddDataset(input storagepool.AddDatasetInput) (storagepool.Pool, storagepool.Dataset, error)
	AddShare(input storagepool.AddShareInput) (storagepool.Share, error)
}

// PoolsLoader assembles the pools response from inventory + declared pool store.
type PoolsLoader struct {
	inventory InventorySource
	store     PoolStore
}

// NewPoolsLoader creates a pools loader. store may be nil (empty declared list).
func NewPoolsLoader(inventory InventorySource, store PoolStore) PoolsLoader {
	return PoolsLoader{inventory: inventory, store: store}
}

// Load returns the current pools DTO including declared pools.
func (l PoolsLoader) Load(ctx context.Context) (PoolsResponse, error) {
	inventory, err := l.inventory.Load(ctx)
	if err != nil {
		return PoolsResponse{}, err
	}
	status := PoolsStatusOK
	if inventory.Status != InventoryStatusOK {
		status = PoolsStatusUnavailable
	}

	pools := []PoolView{}
	if l.store != nil {
		raw, err := l.store.ListPools()
		if err != nil {
			return PoolsResponse{}, err
		}
		pools = mapPools(raw)
	}

	return PoolsResponse{
		CollectedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Status:            status,
		InventoryStatus:   inventory.Status,
		Pools:             pools,
		PoolManagement:    PoolManagementDeclared,
		MutationAvailable: true,
		DiskFormatApplied: false,
		Note:              "Declared pools reserve eligible disks in appliance state. Disk formatting is not applied in this release.",
	}, nil
}

// CreatePoolRequest is the body for POST /api/v1/storage/pools.
type CreatePoolRequest struct {
	Name    string   `json:"name"`
	DiskIDs []string `json:"disk_ids"`
	Confirm bool     `json:"confirm"`
}

// CreatePoolService creates declared pools after eligibility checks.
type CreatePoolService struct {
	inventory InventorySource
	store     PoolStore
}

// NewCreatePoolService wires pool creation.
func NewCreatePoolService(inventory InventorySource, store PoolStore) CreatePoolService {
	return CreatePoolService{inventory: inventory, store: store}
}

// Create validates eligible disks and persists a declared pool.
func (s CreatePoolService) Create(ctx context.Context, req CreatePoolRequest) (PoolView, error) {
	if s.store == nil {
		return PoolView{}, fmt.Errorf("pool store not configured")
	}
	inventory, err := s.inventory.Load(ctx)
	if err != nil {
		return PoolView{}, err
	}
	eligible := map[string]bool{}
	for _, disk := range inventory.Disks {
		if disk.Eligibility == "eligible" {
			eligible[disk.ID] = true
		}
	}
	pool, err := s.store.CreatePool(storagepool.CreatePoolInput{
		Name:    req.Name,
		DiskIDs: req.DiskIDs,
		Confirm: req.Confirm,
	}, eligible)
	if err != nil {
		return PoolView{}, err
	}
	return mapPool(pool), nil
}

// AddDatasetRequest is the body for POST /api/v1/storage/pools/{id}/datasets.
type AddDatasetRequest struct {
	Name string `json:"name"`
}

// AddDatasetService prepares datasets under a declared pool.
type AddDatasetService struct {
	store PoolStore
}

// NewAddDatasetService wires dataset preparation.
func NewAddDatasetService(store PoolStore) AddDatasetService {
	return AddDatasetService{store: store}
}

// Add prepares a dataset plan.
func (s AddDatasetService) Add(poolID string, req AddDatasetRequest) (PoolView, DatasetView, error) {
	if s.store == nil {
		return PoolView{}, DatasetView{}, fmt.Errorf("pool store not configured")
	}
	pool, ds, err := s.store.AddDataset(storagepool.AddDatasetInput{PoolID: poolID, Name: req.Name})
	if err != nil {
		return PoolView{}, DatasetView{}, err
	}
	return mapPool(pool), DatasetView{
		ID: ds.ID, Name: ds.Name, PathIntent: ds.PathIntent, Status: ds.Status, CreatedAt: ds.CreatedAt,
	}, nil
}

func mapPools(raw []storagepool.Pool) []PoolView {
	out := make([]PoolView, 0, len(raw))
	for _, pool := range raw {
		out = append(out, mapPool(pool))
	}
	return out
}

func mapPool(pool storagepool.Pool) PoolView {
	datasets := make([]DatasetView, 0, len(pool.Datasets))
	for _, ds := range pool.Datasets {
		datasets = append(datasets, DatasetView{
			ID: ds.ID, Name: ds.Name, PathIntent: ds.PathIntent, Status: ds.Status, CreatedAt: ds.CreatedAt,
		})
	}
	return PoolView{
		ID: pool.ID, Name: pool.Name, Status: pool.Status,
		DiskIDs: append([]string(nil), pool.DiskIDs...),
		DiskFormatState: pool.DiskFormatState,
		Datasets:        datasets,
		CreatedAt:       pool.CreatedAt,
		UpdatedAt:       pool.UpdatedAt,
	}
}

// MapCreatePoolError maps store errors to HTTP-facing codes.
func MapCreatePoolError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, storagepool.ErrConfirmRequired):
		return 400, "CONFIRM_REQUIRED", "confirm must be true to declare a pool"
	case errors.Is(err, storagepool.ErrInvalidName):
		return 400, "INVALID_NAME", "pool or dataset name is invalid"
	case errors.Is(err, storagepool.ErrNoDisks):
		return 400, "NO_DISKS", "at least one eligible disk is required"
	case errors.Is(err, storagepool.ErrDiskNotEligible):
		return 400, "DISK_NOT_ELIGIBLE", "one or more disks are not eligible"
	case errors.Is(err, storagepool.ErrDiskAlreadyUsed):
		return 409, "DISK_IN_USE", "one or more disks are already assigned to a pool"
	case errors.Is(err, storagepool.ErrPoolNotFound):
		return 404, "POOL_NOT_FOUND", "pool not found"
	case errors.Is(err, storagepool.ErrDuplicateDataset):
		return 409, "DUPLICATE_DATASET", "dataset name already exists on this pool"
	case errors.Is(err, storagepool.ErrDuplicateShare):
		return 409, "DUPLICATE_SHARE", "share name already exists"
	case errors.Is(err, storagepool.ErrDatasetNotFound):
		return 404, "DATASET_NOT_FOUND", "dataset not found on pool"
	default:
		return 500, "INTERNAL_ERROR", "Internal server error"
	}
}
