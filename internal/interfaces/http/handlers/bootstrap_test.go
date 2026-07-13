package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	httpapi "github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/handlers"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/config"
)

func newBootstrapRouter(t *testing.T) (http.Handler, *sqlite.Store) {
	t.Helper()

	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	hasher, err := appidentity.NewPasswordHasher(appidentity.Argon2idConfig{
		Memory:      4096,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() error: %v", err)
	}

	userRepo := sqlite.NewUserRepository(store.DB())
	bootstrapRepo := sqlite.NewBootstrapRepository(store.DB())
	service := appidentity.NewBootstrapService(hasher, bootstrapRepo, userRepo)
	bootstrap := handlers.NewBootstrap(handlers.BootstrapDeps{Service: service})

	cfg := config.Config{
		Version:     "test",
		BuildCommit: "test",
		ReadTimeout: 15 * time.Second,
	}
	router := httpapi.NewRouter(cfg, slog.Default(), health.NewReadiness(store), bootstrap, nil, nil)
	return router, store
}

func bootstrapPOST(remoteAddr, body string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/bootstrap", strings.NewReader(body))
	req.RemoteAddr = remoteAddr
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

func TestBootstrapLoopbackAllowed(t *testing.T) {
	body := `{"username":"owner","password":"valid-password-123"}`

	for _, addr := range []string{"127.0.0.1:8080", "[::1]:8080", "127.0.0.2:1234"} {
		router, _ := newBootstrapRouter(t)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, bootstrapPOST(addr, body, nil))
		if rec.Code != http.StatusCreated {
			t.Fatalf("addr %q status = %d, want 201", addr, rec.Code)
		}
	}
}

func TestBootstrapNonLoopbackRejected(t *testing.T) {
	router, _ := newBootstrapRouter(t)
	body := `{"username":"owner","password":"valid-password-123"}`

	for _, addr := range []string{
		"192.168.1.10:8080",
		"8.8.8.8:443",
		"[fd00::1]:8080",
		"[fe80::1]:8080",
		"",
		"malformed",
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, bootstrapPOST(addr, body, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("addr %q status = %d, want 404", addr, rec.Code)
		}
	}
}

func TestBootstrapForwardedHeadersDoNotBypassGate(t *testing.T) {
	router, _ := newBootstrapRouter(t)
	body := `{"username":"owner","password":"valid-password-123"}`
	headers := map[string]string{
		"X-Forwarded-For":  "127.0.0.1",
		"X-Real-IP":        "127.0.0.1",
		"CF-Connecting-IP": "127.0.0.1",
		"Forwarded":        "for=127.0.0.1",
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, bootstrapPOST("192.168.1.10:8080", body, headers))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestBootstrapCreatesOwnerWithAudit(t *testing.T) {
	router, store := newBootstrapRouter(t)
	body := `{"username":"owner","password":"valid-password-123"}`

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", body, nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["role"] != "owner" || resp["status"] != "active" || resp["username"] != "owner" {
		t.Fatalf("response = %+v", resp)
	}
	for _, forbidden := range []string{"password", "hash", "token", "csrf"} {
		for key := range resp {
			if strings.Contains(strings.ToLower(key), forbidden) {
				t.Fatalf("forbidden response key %q", key)
			}
		}
	}

	userRepo := sqlite.NewUserRepository(store.DB())
	cred, err := userRepo.GetUserByUsername(context.Background(), "owner")
	if err != nil {
		t.Fatalf("GetUserByUsername() error: %v", err)
	}
	if !strings.HasPrefix(cred.PasswordHash, "$argon2id$") {
		t.Fatal("expected argon2id password hash in storage")
	}

	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())
	events, err := auditRepo.ListSecurityAuditEvents(context.Background(), appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "identity.bootstrap.succeeded" {
		t.Fatalf("unexpected audit events")
	}
}

func TestBootstrapUnavailableAfterFirstUser(t *testing.T) {
	router, _ := newBootstrapRouter(t)
	body := `{"username":"owner","password":"valid-password-123"}`

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", body, nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("first bootstrap status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", `{"username":"other","password":"valid-password-456"}`, nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("second bootstrap status = %d, want 404", rec.Code)
	}
}

func TestBootstrapInvalidRequests(t *testing.T) {
	router, _ := newBootstrapRouter(t)

	cases := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "unknown field",
			req:  bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"valid-password-123","role":"owner"}`, nil),
		},
		{
			name: "invalid json",
			req:  bootstrapPOST("127.0.0.1:8080", `{`, nil),
		},
		{
			name: "missing content type",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/bootstrap", strings.NewReader(`{"username":"owner","password":"valid-password-123"}`))
				req.RemoteAddr = "127.0.0.1:8080"
				return req
			}(),
		},
		{
			name: "short password",
			req:  bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"short"}`, nil),
		},
		{
			name: "empty username",
			req:  bootstrapPOST("127.0.0.1:8080", `{"username":"","password":"valid-password-123"}`, nil),
		},
	}

	for _, tc := range cases {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, tc.req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400", tc.name, rec.Code)
		}
	}
}

func TestBootstrapOversizedBodyRejected(t *testing.T) {
	router, _ := newBootstrapRouter(t)
	largePassword := strings.Repeat("a", 9000)
	body := `{"username":"owner","password":"` + largePassword + `"}`

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, bootstrapPOST("127.0.0.1:8080", body, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestBootstrapConcurrentRequestsCreateOneUser(t *testing.T) {
	router, store := newBootstrapRouter(t)
	body := []byte(`{"username":"owner","password":"valid-password-123"}`)

	var wg sync.WaitGroup
	results := make(chan int, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/bootstrap", bytes.NewReader(body))
			req.RemoteAddr = "127.0.0.1:8080"
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			results <- rec.Code
		}()
	}
	wg.Wait()
	close(results)

	created := 0
	for code := range results {
		if code == http.StatusCreated {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("created responses = %d, want 1", created)
	}

	userRepo := sqlite.NewUserRepository(store.DB())
	count, err := userRepo.CountUsers(context.Background())
	if err != nil {
		t.Fatalf("CountUsers() error: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestBootstrapNonLoopbackMatchesUnavailableResponse(t *testing.T) {
	router, _ := newBootstrapRouter(t)

	recExisting := httptest.NewRecorder()
	router.ServeHTTP(recExisting, bootstrapPOST("127.0.0.1:8080", `{"username":"owner","password":"valid-password-123"}`, nil))
	if recExisting.Code != http.StatusCreated {
		t.Fatalf("setup status = %d", recExisting.Code)
	}

	recLAN := httptest.NewRecorder()
	router.ServeHTTP(recLAN, bootstrapPOST("192.168.1.10:8080", `{"username":"other","password":"valid-password-456"}`, nil))

	recBlocked := httptest.NewRecorder()
	router.ServeHTTP(recBlocked, bootstrapPOST("127.0.0.1:8080", `{"username":"other","password":"valid-password-456"}`, nil))

	if recLAN.Code != recBlocked.Code {
		t.Fatalf("LAN status %d != blocked status %d", recLAN.Code, recBlocked.Code)
	}
}
