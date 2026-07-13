package sqlite_test

import (
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestBootstrapRepositoryCreatesFirstOwnerWithAudit(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewBootstrapRepository(store.DB())
	userRepo := sqlite.NewUserRepository(store.DB())
	auditRepo := sqlite.NewSecurityAuditRepository(store.DB())

	created, err := repo.CreateFirstOwner(ctx, appidentity.BootstrapCreateInput{
		UserID:       "owner-1",
		Username:     "owner",
		PasswordHash: "hash-value",
		Role:         "owner",
		Status:       string(appidentity.UserStatusActive),
	}, appidentity.BootstrapAuditInput{
		ID:            "audit-1",
		ActorUserID:   "owner-1",
		SubjectUserID: "owner-1",
		EventType:     "identity.bootstrap.succeeded",
		Result:        "success",
		MetadataJSON:  `{"source":"loopback"}`,
	})
	if err != nil {
		t.Fatalf("CreateFirstOwner() error: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}

	count, err := userRepo.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers() error: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	events, err := auditRepo.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "identity.bootstrap.succeeded" {
		t.Fatalf("unexpected audit events: %+v", events)
	}

	created, err = repo.CreateFirstOwner(ctx, appidentity.BootstrapCreateInput{
		UserID:       "owner-2",
		Username:     "other",
		PasswordHash: "hash-value",
		Role:         "owner",
		Status:       string(appidentity.UserStatusActive),
	}, appidentity.BootstrapAuditInput{
		ID:            "audit-2",
		ActorUserID:   "owner-2",
		SubjectUserID: "owner-2",
		EventType:     "identity.bootstrap.succeeded",
		Result:        "success",
	})
	if err != nil {
		t.Fatalf("second CreateFirstOwner() error: %v", err)
	}
	if created {
		t.Fatal("expected second create to fail")
	}
}

func TestBootstrapRepositoryConcurrentCreateOnlyOneUser(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewBootstrapRepository(store.DB())
	userRepo := sqlite.NewUserRepository(store.DB())

	type result struct {
		created bool
		err     error
	}
	results := make(chan result, 2)

	for i := 0; i < 2; i++ {
		go func(idx int) {
			created, err := repo.CreateFirstOwner(ctx, appidentity.BootstrapCreateInput{
				UserID:       "owner-" + string(rune('a'+idx)),
				Username:     "owner" + string(rune('a'+idx)),
				PasswordHash: "hash-value",
				Role:         "owner",
				Status:       string(appidentity.UserStatusActive),
			}, appidentity.BootstrapAuditInput{
				ID:            "audit-" + string(rune('a'+idx)),
				ActorUserID:   "owner-" + string(rune('a'+idx)),
				SubjectUserID: "owner-" + string(rune('a'+idx)),
				EventType:     "identity.bootstrap.succeeded",
				Result:        "success",
			})
			results <- result{created: created, err: err}
		}(i)
	}

	successes := 0
	for i := 0; i < 2; i++ {
		res := <-results
		if res.err != nil {
			t.Fatalf("CreateFirstOwner() error: %v", res.err)
		}
		if res.created {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successes = %d, want 1", successes)
	}

	count, err := userRepo.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers() error: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}
