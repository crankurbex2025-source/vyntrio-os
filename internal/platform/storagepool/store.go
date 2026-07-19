package storagepool

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	schemaVersion = "vyntrio-storage-pools-v1"
	storeFileName = "pools.json"

	StatusDeclared = "declared"
	StatusPlanned  = "planned"

	// DiskFormatPending means disks are reserved in state but not formatted.
	DiskFormatPending = "pending"
)

var (
	ErrNotFound          = errors.New("storagepool: not found")
	ErrInvalidName       = errors.New("storagepool: invalid name")
	ErrNoDisks           = errors.New("storagepool: at least one disk is required")
	ErrDiskNotEligible   = errors.New("storagepool: disk is not eligible")
	ErrDiskAlreadyUsed   = errors.New("storagepool: disk already assigned to a pool")
	ErrConfirmRequired   = errors.New("storagepool: confirm must be true")
	ErrDuplicateDataset  = errors.New("storagepool: dataset name already exists")
	ErrDuplicateShare    = errors.New("storagepool: share name already exists")
	ErrPoolNotFound      = errors.New("storagepool: pool not found")
	ErrDatasetNotFound   = errors.New("storagepool: dataset not found")

	namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
)

// Pool is a declared storage pool. Disks are reserved in appliance state;
// formatting and filesystem creation are not applied in this foundation slice.
type Pool struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	DiskIDs         []string  `json:"disk_ids"`
	DiskFormatState string    `json:"disk_format_state"`
	Datasets        []Dataset `json:"datasets"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

// Dataset is a planned dataset/volume layout under a declared pool.
type Dataset struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PathIntent string `json:"path_intent"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// Share is a planned network share (no SMB/NFS daemon yet).
type Share struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PoolID    string `json:"pool_id"`
	DatasetID string `json:"dataset_id,omitempty"`
	Protocol  string `json:"protocol"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type storeDocument struct {
	SchemaVersion string  `json:"schema_version"`
	Pools         []Pool  `json:"pools"`
	Shares        []Share `json:"shares"`
}

// Store persists declared pools and share plans under the appliance state directory.
type Store struct {
	mu       sync.Mutex
	stateDir string
}

// NewStore creates a store rooted at stateDir/storage/.
func NewStore(stateDir string) *Store {
	return &Store{stateDir: strings.TrimSpace(stateDir)}
}

func (s *Store) path() string {
	return filepath.Join(s.stateDir, "storage", storeFileName)
}

func (s *Store) load() (storeDocument, error) {
	path := s.path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return storeDocument{SchemaVersion: schemaVersion, Pools: []Pool{}, Shares: []Share{}}, nil
		}
		return storeDocument{}, err
	}
	var doc storeDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return storeDocument{}, fmt.Errorf("decode store: %w", err)
	}
	if doc.Pools == nil {
		doc.Pools = []Pool{}
	}
	if doc.Shares == nil {
		doc.Shares = []Share{}
	}
	doc.SchemaVersion = schemaVersion
	return doc, nil
}

func (s *Store) save(doc storeDocument) error {
	dir := filepath.Dir(s.path())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	doc.SchemaVersion = schemaVersion
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o640); err != nil {
		return err
	}
	return os.Rename(tmp, s.path())
}

// ListPools returns all declared pools.
func (s *Store) ListPools() ([]Pool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]Pool, len(doc.Pools))
	copy(out, doc.Pools)
	return out, nil
}

// ListShares returns all planned shares.
func (s *Store) ListShares() ([]Share, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]Share, len(doc.Shares))
	copy(out, doc.Shares)
	return out, nil
}

// UsedDiskIDs returns disk IDs already assigned to any pool.
func (s *Store) UsedDiskIDs() (map[string]string, error) {
	pools, err := s.ListPools()
	if err != nil {
		return nil, err
	}
	used := make(map[string]string)
	for _, pool := range pools {
		for _, id := range pool.DiskIDs {
			used[id] = pool.ID
		}
	}
	return used, nil
}

// CreatePoolInput creates a declared pool.
type CreatePoolInput struct {
	Name    string
	DiskIDs []string
	Confirm bool
	Now     time.Time
	NewID   func() string
}

// CreatePool validates and persists a declared pool without formatting disks.
func (s *Store) CreatePool(input CreatePoolInput, eligible map[string]bool) (Pool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !input.Confirm {
		return Pool{}, ErrConfirmRequired
	}
	name := strings.TrimSpace(strings.ToLower(input.Name))
	if !namePattern.MatchString(name) {
		return Pool{}, ErrInvalidName
	}
	if len(input.DiskIDs) == 0 {
		return Pool{}, ErrNoDisks
	}

	doc, err := s.load()
	if err != nil {
		return Pool{}, err
	}

	used := make(map[string]struct{})
	for _, pool := range doc.Pools {
		if pool.Name == name {
			return Pool{}, ErrInvalidName
		}
		for _, id := range pool.DiskIDs {
			used[id] = struct{}{}
		}
	}

	diskIDs := make([]string, 0, len(input.DiskIDs))
	seen := make(map[string]struct{})
	for _, raw := range input.DiskIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if !eligible[id] {
			return Pool{}, fmt.Errorf("%w: %s", ErrDiskNotEligible, id)
		}
		if _, ok := used[id]; ok {
			return Pool{}, fmt.Errorf("%w: %s", ErrDiskAlreadyUsed, id)
		}
		diskIDs = append(diskIDs, id)
	}
	if len(diskIDs) == 0 {
		return Pool{}, ErrNoDisks
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	idFn := input.NewID
	if idFn == nil {
		idFn = func() string { return fmt.Sprintf("pool-%d", now.UnixNano()) }
	}
	pool := Pool{
		ID:              idFn(),
		Name:            name,
		Status:          StatusDeclared,
		DiskIDs:         diskIDs,
		DiskFormatState: DiskFormatPending,
		Datasets:        []Dataset{},
		CreatedAt:       now.UTC().Format(time.RFC3339Nano),
		UpdatedAt:       now.UTC().Format(time.RFC3339Nano),
	}
	doc.Pools = append(doc.Pools, pool)
	if err := s.save(doc); err != nil {
		return Pool{}, err
	}
	return pool, nil
}

// AddDatasetInput prepares a dataset under a declared pool.
type AddDatasetInput struct {
	PoolID string
	Name   string
	Now    time.Time
	NewID  func() string
}

// AddDataset appends a planned dataset to a pool.
func (s *Store) AddDataset(input AddDatasetInput) (Pool, Dataset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := strings.TrimSpace(strings.ToLower(input.Name))
	if !namePattern.MatchString(name) {
		return Pool{}, Dataset{}, ErrInvalidName
	}
	doc, err := s.load()
	if err != nil {
		return Pool{}, Dataset{}, err
	}
	idx := -1
	for i := range doc.Pools {
		if doc.Pools[i].ID == input.PoolID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Pool{}, Dataset{}, ErrPoolNotFound
	}
	for _, ds := range doc.Pools[idx].Datasets {
		if ds.Name == name {
			return Pool{}, Dataset{}, ErrDuplicateDataset
		}
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	idFn := input.NewID
	if idFn == nil {
		idFn = func() string { return fmt.Sprintf("ds-%d", now.UnixNano()) }
	}
	ds := Dataset{
		ID:         idFn(),
		Name:       name,
		PathIntent: filepath.Join("/", doc.Pools[idx].Name, name),
		Status:     StatusPlanned,
		CreatedAt:  now.UTC().Format(time.RFC3339Nano),
	}
	doc.Pools[idx].Datasets = append(doc.Pools[idx].Datasets, ds)
	doc.Pools[idx].UpdatedAt = now.UTC().Format(time.RFC3339Nano)
	if err := s.save(doc); err != nil {
		return Pool{}, Dataset{}, err
	}
	return doc.Pools[idx], ds, nil
}

// AddShareInput prepares a share plan.
type AddShareInput struct {
	Name      string
	PoolID    string
	DatasetID string
	Protocol  string
	Now       time.Time
	NewID     func() string
}

// AddShare appends a planned share (protocol remains planned until SMB/NFS ships).
func (s *Store) AddShare(input AddShareInput) (Share, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := strings.TrimSpace(strings.ToLower(input.Name))
	if !namePattern.MatchString(name) {
		return Share{}, ErrInvalidName
	}
	protocol := strings.TrimSpace(strings.ToLower(input.Protocol))
	if protocol == "" {
		protocol = "planned"
	}
	if protocol != "planned" && protocol != "smb" && protocol != "nfs" {
		return Share{}, ErrInvalidName
	}
	// Until protocol daemons exist, always persist as planned.
	protocol = "planned"

	doc, err := s.load()
	if err != nil {
		return Share{}, err
	}
	var pool *Pool
	for i := range doc.Pools {
		if doc.Pools[i].ID == input.PoolID {
			pool = &doc.Pools[i]
			break
		}
	}
	if pool == nil {
		return Share{}, ErrPoolNotFound
	}
	if input.DatasetID != "" {
		found := false
		for _, ds := range pool.Datasets {
			if ds.ID == input.DatasetID {
				found = true
				break
			}
		}
		if !found {
			return Share{}, ErrDatasetNotFound
		}
	}
	for _, share := range doc.Shares {
		if share.Name == name {
			return Share{}, ErrDuplicateShare
		}
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	idFn := input.NewID
	if idFn == nil {
		idFn = func() string { return fmt.Sprintf("share-%d", now.UnixNano()) }
	}
	share := Share{
		ID:        idFn(),
		Name:      name,
		PoolID:    input.PoolID,
		DatasetID: strings.TrimSpace(input.DatasetID),
		Protocol:  protocol,
		Status:    StatusPlanned,
		CreatedAt: now.UTC().Format(time.RFC3339Nano),
	}
	doc.Shares = append(doc.Shares, share)
	if err := s.save(doc); err != nil {
		return Share{}, err
	}
	return share, nil
}
