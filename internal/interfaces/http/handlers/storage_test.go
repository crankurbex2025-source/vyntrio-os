package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appstorage "github.com/crankurbex2025-source/vyntrio-os/internal/application/storage"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

const storageDisksPath = "/api/v1/storage/disks"

type storageDisksResponse struct {
	CollectedAt string `json:"collected_at"`
	Status      string `json:"status"`
	Disks       []struct {
		ID          string   `json:"id"`
		Status      string   `json:"status"`
		SizeBytes   *uint64  `json:"size_bytes,omitempty"`
		Rotational  *bool    `json:"rotational,omitempty"`
		Removable   *bool    `json:"removable,omitempty"`
		Eligibility string   `json:"eligibility"`
		Reasons     []string `json:"reasons,omitempty"`
	} `json:"disks"`
}

func storageDisksGET(cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, storageDisksPath, nil)
	req.RemoteAddr = "127.0.0.1:8080"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func TestStorageDisksUnauthenticatedReturns401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storageDisksGET(nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestStorageDisksMissingPermissionReturns403(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{authorizer: denyAuthorizer{}})

	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storageDisksGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestStorageDisksRBACMatrix(t *testing.T) {
	size := uint64(1_000_000_000_000)
	rot := false
	reader := stubBlockDeviceReader{
		devices: []storageinventory.RawDevice{
			{
				KernelName: "sdb",
				SizeBytes:  size,
				SizeKnown:  true,
				Rotational: &rot,
			},
		},
	}
	router := newSettingsRouter(t, settingsRouterOpts{storageReader: reader})

	cases := []struct {
		name       string
		userID     domainidentity.UserID
		username   string
		role       domainidentity.Role
		sessionID  string
		rawToken   string
		wantStatus int
	}{
		{
			name:       "owner",
			userID:     domainidentity.UserID("storage-owner"),
			username:   "storage-owner",
			role:       domainidentity.RoleOwner,
			sessionID:  "storage-owner-sess",
			rawToken:   "storage-owner-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "operator",
			userID:     domainidentity.UserID("storage-operator"),
			username:   "storage-operator",
			role:       domainidentity.RoleOperator,
			sessionID:  "storage-operator-sess",
			rawToken:   "storage-operator-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "read_only",
			userID:     domainidentity.UserID("storage-read-only"),
			username:   "storage-read-only",
			role:       domainidentity.RoleReadOnly,
			sessionID:  "storage-read-only-sess",
			rawToken:   "storage-read-only-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionCookie := createRoleSettingsSession(
				t,
				router,
				tc.userID,
				tc.username,
				tc.role,
				tc.sessionID,
				tc.rawToken,
			)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, storageDisksGET([]*http.Cookie{sessionCookie}))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
			if tc.wantStatus != http.StatusOK {
				return
			}

			assertStorageDisksResponseShape(t, rec)
			assertStorageCacheControlNoStore(t, rec)
			assertNoKernelNamesInStorageResponse(t, rec.Body.Bytes())
		})
	}
}

func TestStorageDisksEligibleCandidateSerialization(t *testing.T) {
	size := uint64(2_000_000_000_000)
	rot := true
	reader := stubBlockDeviceReader{
		devices: []storageinventory.RawDevice{
			{
				KernelName: "sdb",
				SizeBytes:  size,
				SizeKnown:  true,
				Rotational: &rot,
			},
		},
	}
	router := newSettingsRouter(t, settingsRouterOpts{storageReader: reader})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storageDisksGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	got := decodeStorageDisksResponse(t, rec.Body.Bytes())
	if got.Status != storageinventory.StatusOK {
		t.Fatalf("status = %q", got.Status)
	}
	if len(got.Disks) != 1 {
		t.Fatalf("len(disks) = %d, want 1", len(got.Disks))
	}
	disk := got.Disks[0]
	if disk.Eligibility != storageinventory.EligibilityEligible {
		t.Fatalf("eligibility = %q", disk.Eligibility)
	}
	if disk.SizeBytes == nil || *disk.SizeBytes != size {
		t.Fatalf("size_bytes = %v", disk.SizeBytes)
	}
	if disk.Rotational == nil || *disk.Rotational != rot {
		t.Fatalf("rotational = %v", disk.Rotational)
	}
	if !strings.HasPrefix(disk.ID, "disk-") {
		t.Fatalf("id = %q, want opaque disk- prefix", disk.ID)
	}
}

func TestStorageDisksReaderFailureReturnsUnavailable(t *testing.T) {
	reader := stubBlockDeviceReader{err: errStorageDiscoveryFailed}
	router := newSettingsRouter(t, settingsRouterOpts{storageReader: reader})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storageDisksGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	got := decodeStorageDisksResponse(t, rec.Body.Bytes())
	if got.Status != storageinventory.StatusUnavailable {
		t.Fatalf("status = %q, want unavailable", got.Status)
	}
	if len(got.Disks) != 0 {
		t.Fatalf("disks = %v, want empty", got.Disks)
	}
}

type stubBlockDeviceReader struct {
	devices []storageinventory.RawDevice
	err     error
}

var errStorageDiscoveryFailed = &storageDiscoveryError{}

type storageDiscoveryError struct{}

func (storageDiscoveryError) Error() string { return "discovery failed" }

func (s stubBlockDeviceReader) ListBlockDevices(string) ([]storageinventory.RawDevice, error) {
	return s.devices, s.err
}

func decodeStorageDisksResponse(t *testing.T, body []byte) storageDisksResponse {
	t.Helper()
	var got storageDisksResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v body=%s", err, string(body))
	}
	return got
}

func assertStorageDisksResponseShape(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	got := decodeStorageDisksResponse(t, rec.Body.Bytes())
	if got.CollectedAt == "" {
		t.Fatal("expected collected_at")
	}
	if got.Status == "" {
		t.Fatal("expected status")
	}
	if got.Disks == nil {
		t.Fatal("expected disks array")
	}
}

func assertStorageCacheControlNoStore(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func assertNoKernelNamesInStorageResponse(t *testing.T, body []byte) {
	t.Helper()
	raw := string(body)
	for _, forbidden := range []string{`"sda"`, `"/dev/`, `"nvme`, `"kernel"`} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("response leaks device identity: contains %q", forbidden)
		}
	}
}

func TestStorageLoaderMapsInventory(t *testing.T) {
	size := uint64(500_000_000_000)
	collector := storageinventory.NewCollector("/var/lib/vyntrio", storageinventory.CollectorDeps{
		Reader: stubBlockDeviceReader{
			devices: []storageinventory.RawDevice{
				{
					KernelName: "sdb",
					SizeBytes:  size,
					SizeKnown:  true,
				},
			},
		},
	})
	loader := appstorage.NewInventoryLoader(collector)
	got, err := loader.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Status != storageinventory.StatusOK {
		t.Fatalf("status = %q", got.Status)
	}
	if len(got.Disks) != 1 {
		t.Fatalf("len(disks) = %d", len(got.Disks))
	}
}

const storagePoolsPath = "/api/v1/storage/pools"
const storageSharesPath = "/api/v1/storage/shares"

type storagePoolsResponse struct {
	CollectedAt       string `json:"collected_at"`
	Status            string `json:"status"`
	InventoryStatus   string `json:"inventory_status"`
	Pools             []any  `json:"pools"`
	PoolManagement    string `json:"pool_management"`
	MutationAvailable bool   `json:"mutation_available"`
}

type storageSharesResponse struct {
	CollectedAt       string `json:"collected_at"`
	Status            string `json:"status"`
	InventoryStatus   string `json:"inventory_status"`
	Shares            []any  `json:"shares"`
	ShareManagement   string `json:"share_management"`
	ProtocolSupport   string `json:"protocol_support"`
	MutationAvailable bool   `json:"mutation_available"`
}

func storagePoolsGET(cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, storagePoolsPath, nil)
	req.RemoteAddr = "127.0.0.1:8080"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func storageSharesGET(cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, storageSharesPath, nil)
	req.RemoteAddr = "127.0.0.1:8080"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func TestStoragePoolsUnauthenticatedReturns401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storagePoolsGET(nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestStorageSharesUnauthenticatedReturns401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storageSharesGET(nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestStoragePoolsEmptyHonestResponse(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storagePoolsGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var got storagePoolsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if got.Status != appstorage.PoolsStatusOK || len(got.Pools) != 0 {
		t.Fatalf("response = %+v", got)
	}
	if got.PoolManagement != appstorage.PoolManagementDeclared || !got.MutationAvailable {
		t.Fatalf("pool_management/mutation = %q %v", got.PoolManagement, got.MutationAvailable)
	}
	assertStorageCacheControlNoStore(t, rec)
}

func TestStorageSharesEmptyHonestResponse(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, storageSharesGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var got storageSharesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if got.Status != appstorage.SharesStatusOK || len(got.Shares) != 0 {
		t.Fatalf("response = %+v", got)
	}
	if got.ShareManagement != appstorage.ShareManagementPlanned || got.ProtocolSupport != "not_available" || !got.MutationAvailable {
		t.Fatalf("share_management/protocol/mutation = %q %q %v", got.ShareManagement, got.ProtocolSupport, got.MutationAvailable)
	}
	assertStorageCacheControlNoStore(t, rec)
}

