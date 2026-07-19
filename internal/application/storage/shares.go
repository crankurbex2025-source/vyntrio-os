package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storagepool"
)

const (
	SharesStatusOK              = "ok"
	SharesStatusUnavailable     = "unavailable"
	ShareManagementPlanned      = "planned_shares"
	ShareManagementNotAvailable = "not_available"
)

// ShareView is one planned share in GET /api/v1/storage/shares.
type ShareView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PoolID    string `json:"pool_id"`
	DatasetID string `json:"dataset_id,omitempty"`
	Protocol  string `json:"protocol"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// SharesResponse is the network-share view for GET /api/v1/storage/shares.
type SharesResponse struct {
	CollectedAt       string      `json:"collected_at"`
	Status            string      `json:"status"`
	InventoryStatus   string      `json:"inventory_status"`
	Shares            []ShareView `json:"shares"`
	ShareManagement   string      `json:"share_management"`
	ProtocolSupport   string      `json:"protocol_support"`
	MutationAvailable bool        `json:"mutation_available"`
	Note              string      `json:"note,omitempty"`
}

// SharesLoader assembles shares from inventory health + planned share store.
type SharesLoader struct {
	inventory InventorySource
	store     PoolStore
}

// NewSharesLoader creates a shares loader.
func NewSharesLoader(inventory InventorySource, store PoolStore) SharesLoader {
	return SharesLoader{inventory: inventory, store: store}
}

// Load returns planned shares (no SMB/NFS daemon yet).
func (l SharesLoader) Load(ctx context.Context) (SharesResponse, error) {
	inventory, err := l.inventory.Load(ctx)
	if err != nil {
		return SharesResponse{}, err
	}
	status := SharesStatusOK
	if inventory.Status != InventoryStatusOK {
		status = SharesStatusUnavailable
	}
	shares := []ShareView{}
	if l.store != nil {
		raw, err := l.store.ListShares()
		if err != nil {
			return SharesResponse{}, err
		}
		shares = mapShares(raw)
	}
	return SharesResponse{
		CollectedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Status:            status,
		InventoryStatus:   inventory.Status,
		Shares:            shares,
		ShareManagement:   ShareManagementPlanned,
		ProtocolSupport:   "not_available",
		MutationAvailable: true,
		Note:              "Share plans are stored in appliance state. SMB/NFS protocol services are not available yet.",
	}, nil
}

// CreateShareRequest is the body for POST /api/v1/storage/shares.
type CreateShareRequest struct {
	Name      string `json:"name"`
	PoolID    string `json:"pool_id"`
	DatasetID string `json:"dataset_id,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
}

// CreateShareService prepares share plans.
type CreateShareService struct {
	store PoolStore
}

// NewCreateShareService wires share preparation.
func NewCreateShareService(store PoolStore) CreateShareService {
	return CreateShareService{store: store}
}

// Create persists a planned share.
func (s CreateShareService) Create(req CreateShareRequest) (ShareView, error) {
	if s.store == nil {
		return ShareView{}, fmt.Errorf("pool store not configured")
	}
	share, err := s.store.AddShare(storagepool.AddShareInput{
		Name: req.Name, PoolID: req.PoolID, DatasetID: req.DatasetID, Protocol: req.Protocol,
	})
	if err != nil {
		return ShareView{}, err
	}
	return mapShare(share), nil
}

func mapShares(raw []storagepool.Share) []ShareView {
	out := make([]ShareView, 0, len(raw))
	for _, share := range raw {
		out = append(out, mapShare(share))
	}
	return out
}

func mapShare(share storagepool.Share) ShareView {
	return ShareView{
		ID: share.ID, Name: share.Name, PoolID: share.PoolID,
		DatasetID: share.DatasetID, Protocol: share.Protocol,
		Status: share.Status, CreatedAt: share.CreatedAt,
	}
}
