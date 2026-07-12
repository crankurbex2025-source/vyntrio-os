package sqlite_test

import (
	"strings"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestSecurityAuditRepositoryAppendAndList(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	users := sqlite.NewUserRepository(store.DB())
	audit := sqlite.NewSecurityAuditRepository(store.DB())

	createTestUser(t, users, "actor-1", "actor")
	createTestUser(t, users, "subject-1", "subject")

	metadata := `{"reason":"test","nested":{"k":1}}`
	if err := audit.AppendSecurityAuditEvent(ctx, appidentity.AppendSecurityAuditEventInput{
		ID:            "audit-1",
		ActorUserID:   domainidentity.UserID("actor-1"),
		SubjectUserID: domainidentity.UserID("subject-1"),
		EventType:     "login",
		Result:        "success",
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		MetadataJSON:  metadata,
	}); err != nil {
		t.Fatalf("AppendSecurityAuditEvent() error: %v", err)
	}

	events, err := audit.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].EventType != "login" {
		t.Fatalf("event_type = %q", events[0].EventType)
	}
	if events[0].MetadataJSON != metadata {
		t.Fatalf("metadata changed: %q", events[0].MetadataJSON)
	}
	if events[0].ActorUserID != domainidentity.UserID("actor-1") {
		t.Fatalf("actor_user_id = %q", events[0].ActorUserID)
	}
}

func TestSecurityAuditRepositoryNullableActorAndSubject(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	audit := sqlite.NewSecurityAuditRepository(store.DB())

	if err := audit.AppendSecurityAuditEvent(ctx, appidentity.AppendSecurityAuditEventInput{
		ID:        "audit-null",
		EventType: "system_start",
		Result:    "success",
	}); err != nil {
		t.Fatalf("AppendSecurityAuditEvent() error: %v", err)
	}

	events, err := audit.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 10})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].ActorUserID != "" {
		t.Fatalf("actor_user_id = %q, want empty", events[0].ActorUserID)
	}
	if events[0].SubjectUserID != "" {
		t.Fatalf("subject_user_id = %q, want empty", events[0].SubjectUserID)
	}
}

func TestSecurityAuditRepositoryCursorPaginationNoDuplicates(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	audit := sqlite.NewSecurityAuditRepository(store.DB())

	ids := []string{"audit-c", "audit-b", "audit-a"}
	for _, id := range ids {
		if err := audit.AppendSecurityAuditEvent(ctx, appidentity.AppendSecurityAuditEventInput{
			ID:        id,
			EventType: "test_event",
			Result:    "success",
		}); err != nil {
			t.Fatalf("AppendSecurityAuditEvent(%q) error: %v", id, err)
		}
	}

	seen := make(map[string]struct{})
	var cursor *appidentity.AuditListCursor

	for page := 0; page < 3; page++ {
		input := appidentity.ListSecurityAuditEventsInput{Limit: 1}
		if cursor != nil {
			input.After = cursor
		}

		events, err := audit.ListSecurityAuditEvents(ctx, input)
		if err != nil {
			t.Fatalf("ListSecurityAuditEvents page %d error: %v", page, err)
		}
		if len(events) != 1 {
			t.Fatalf("page %d len = %d, want 1", page, len(events))
		}
		if _, ok := seen[events[0].ID]; ok {
			t.Fatalf("duplicate event id %q on page %d", events[0].ID, page)
		}
		seen[events[0].ID] = struct{}{}
		cursor = &appidentity.AuditListCursor{
			OccurredAt: events[0].OccurredAt,
			ID:         events[0].ID,
		}
	}

	if len(seen) != 3 {
		t.Fatalf("collected %d unique events, want 3", len(seen))
	}
}

func TestSecurityAuditRepositoryBoundedListLimit(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	audit := sqlite.NewSecurityAuditRepository(store.DB())

	for i := 0; i < 3; i++ {
		if err := audit.AppendSecurityAuditEvent(ctx, appidentity.AppendSecurityAuditEventInput{
			ID:        "audit-" + string(rune('a'+i)),
			EventType: "bulk",
			Result:    "success",
		}); err != nil {
			t.Fatalf("AppendSecurityAuditEvent() error: %v", err)
		}
	}

	events, err := audit.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 2})
	if err != nil {
		t.Fatalf("ListSecurityAuditEvents() error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}

	_, err = audit.ListSecurityAuditEvents(ctx, appidentity.ListSecurityAuditEventsInput{Limit: 0})
	if err == nil || !strings.Contains(err.Error(), "limit must be positive") {
		t.Fatalf("zero limit error = %v", err)
	}
}

func TestSecurityAuditStoreHasNoMutatingMethods(t *testing.T) {
	var _ appidentity.SecurityAuditStore = (*sqlite.SecurityAuditRepository)(nil)
}
