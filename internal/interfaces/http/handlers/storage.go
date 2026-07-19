package handlers

import (
	"encoding/json"
	"net/http"

	appstorage "github.com/crankurbex2025-source/vyntrio-os/internal/application/storage"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/middleware"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/response"
	"github.com/go-chi/chi/v5"
)

// Storage serves storage layout endpoints under /api/v1/storage/*.
type Storage struct {
	disks         appstorage.InventoryLoader
	pools         appstorage.PoolsLoader
	shares        appstorage.SharesLoader
	createPool    *appstorage.CreatePoolService
	addDataset    *appstorage.AddDatasetService
	createShare   *appstorage.CreateShareService
}

// StorageDeps configures the storage handler.
type StorageDeps struct {
	Loader            appstorage.InventoryLoader
	PoolsLoader       appstorage.PoolsLoader
	SharesLoader      appstorage.SharesLoader
	CreatePoolService *appstorage.CreatePoolService
	AddDatasetService *appstorage.AddDatasetService
	CreateShareService *appstorage.CreateShareService
}

// NewStorage creates storage handlers.
func NewStorage(deps StorageDeps) *Storage {
	return &Storage{
		disks:       deps.Loader,
		pools:       deps.PoolsLoader,
		shares:      deps.SharesLoader,
		createPool:  deps.CreatePoolService,
		addDataset:  deps.AddDatasetService,
		createShare: deps.CreateShareService,
	}
}

// ServeDisks handles GET /api/v1/storage/disks.
func (h *Storage) ServeDisks(w http.ResponseWriter, r *http.Request) {
	h.serveStorageJSON(w, r, func() (any, error) {
		return h.disks.Load(r.Context())
	})
}

// ServePools handles GET /api/v1/storage/pools.
func (h *Storage) ServePools(w http.ResponseWriter, r *http.Request) {
	h.serveStorageJSON(w, r, func() (any, error) {
		return h.pools.Load(r.Context())
	})
}

// ServeShares handles GET /api/v1/storage/shares.
func (h *Storage) ServeShares(w http.ResponseWriter, r *http.Request) {
	h.serveStorageJSON(w, r, func() (any, error) {
		return h.shares.Load(r.Context())
	})
}

// ServeCreatePool handles POST /api/v1/storage/pools.
func (h *Storage) ServeCreatePool(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	w.Header().Set("Cache-Control", "no-store")
	if h.createPool == nil {
		response.Error(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Pool creation is not configured", requestID)
		return
	}
	var req appstorage.CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", requestID)
		return
	}
	pool, err := h.createPool.Create(r.Context(), req)
	if err != nil {
		status, code, message := appstorage.MapCreatePoolError(err)
		response.Error(w, status, code, message, requestID)
		return
	}
	response.JSON(w, http.StatusCreated, pool)
}

// ServeAddDataset handles POST /api/v1/storage/pools/{poolID}/datasets.
func (h *Storage) ServeAddDataset(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	w.Header().Set("Cache-Control", "no-store")
	if h.addDataset == nil {
		response.Error(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Dataset preparation is not configured", requestID)
		return
	}
	poolID := chi.URLParam(r, "poolID")
	var req appstorage.AddDatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", requestID)
		return
	}
	pool, dataset, err := h.addDataset.Add(poolID, req)
	if err != nil {
		status, code, message := appstorage.MapCreatePoolError(err)
		response.Error(w, status, code, message, requestID)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"pool": pool, "dataset": dataset})
}

// ServeCreateShare handles POST /api/v1/storage/shares.
func (h *Storage) ServeCreateShare(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	w.Header().Set("Cache-Control", "no-store")
	if h.createShare == nil {
		response.Error(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Share preparation is not configured", requestID)
		return
	}
	var req appstorage.CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", requestID)
		return
	}
	share, err := h.createShare.Create(req)
	if err != nil {
		status, code, message := appstorage.MapCreatePoolError(err)
		response.Error(w, status, code, message, requestID)
		return
	}
	response.JSON(w, http.StatusCreated, share)
}

// ServeHTTP implements GET /api/v1/storage/disks for backward compatibility.
func (h *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeDisks(w, r)
}

func (h *Storage) serveStorageJSON(w http.ResponseWriter, r *http.Request, load func() (any, error)) {
	requestID := middleware.GetRequestID(r.Context())

	w.Header().Set("Cache-Control", "no-store")

	payload, err := load()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	response.JSON(w, http.StatusOK, payload)
}
