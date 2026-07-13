package sqlite_test

import (
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestInstanceDisplayNameRepositoryUpdatesWithAudit(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	createTestUser(t, users, "owner-1", "owner")

	repo := sqlite.NewInstanceDisplayNameRepository(store.DB())
	changed, err := repo.UpdateInstanceDisplayNameWithAudit(ctx, "Updated Name", domainidentity.UserID("owner-1"), "audit-1")
	if err != nil {
		t.Fatalf("UpdateInstanceDisplayNameWithAudit() error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	settingsRepo := sqlite.NewSettingsRepository(store.DB())
	host, err := settingsRepo.Get(ctx, "system", "hostname")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if host.Value != "Updated Name" {
		t.Fatalf("hostname = %q", host.Value)
	}

	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())
	events, err := auditRepo.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 || events[0].EventType != appsettings.AuditEventInstanceDisplayNameUpdated {
		t.Fatalf("events = %+v", events)
	}
	if events[0].MetadataJSON != "{}" {
		t.Fatalf("metadata = %q", events[0].MetadataJSON)
	}
}

func TestInstanceDisplayNameRepositoryNoOpSkipsAudit(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	createTestUser(t, users, "owner-1", "owner")

	repo := sqlite.NewInstanceDisplayNameRepository(store.DB())
	changed, err := repo.UpdateInstanceDisplayNameWithAudit(ctx, "vyntrio", domainidentity.UserID("owner-1"), "audit-1")
	if err != nil {
		t.Fatalf("UpdateInstanceDisplayNameWithAudit() error: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false for seed hostname")
	}

	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())
	events, err := auditRepo.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events = %+v", events)
	}
}
