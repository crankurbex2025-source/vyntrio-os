package sqlite_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func openIdentityTestDB(t *testing.T) (*sqlite.Store, context.Context) {
	t.Helper()

	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, context.Background()
}

func createTestUser(t *testing.T, repo *sqlite.UserRepository, id, username string) {
	t.Helper()

	if err := repo.CreateUser(context.Background(), appidentity.CreateUserInput{
		ID:                 domainidentity.UserID(id),
		Username:           username,
		DisplayName:        "Display " + username,
		PasswordHash:       "hash-" + username,
		Role:               domainidentity.RoleAdministrator,
		Status:             appidentity.UserStatusActive,
		MustChangePassword: false,
	}); err != nil {
		t.Fatalf("CreateUser(%q) error: %v", username, err)
	}
}

func TestUserRepositoryCreateAndGetByID(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	createTestUser(t, repo, "user-1", "alice")

	got, err := repo.GetUserByID(ctx, domainidentity.UserID("user-1"))
	if err != nil {
		t.Fatalf("GetUserByID() error: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("username = %q, want alice", got.Username)
	}
	if got.DisplayName != "Display alice" {
		t.Fatalf("display_name = %q", got.DisplayName)
	}
	if got.Role != domainidentity.RoleAdministrator {
		t.Fatalf("role = %q", got.Role)
	}
	if got.Status != appidentity.UserStatusActive {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestUserRepositoryGetByIDNotFound(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	_, err := repo.GetUserByID(ctx, domainidentity.UserID("missing"))
	if !errors.Is(err, appidentity.ErrNotFound) {
		t.Fatalf("GetUserByID() error = %v, want ErrNotFound", err)
	}
}

func TestUserRepositoryGetByUsernameCaseInsensitive(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	createTestUser(t, repo, "user-1", "AdminUser")

	cred, err := repo.GetUserByUsername(ctx, "adminuser")
	if err != nil {
		t.Fatalf("GetUserByUsername() error: %v", err)
	}
	if cred.User.Username != "AdminUser" {
		t.Fatalf("username = %q, want AdminUser", cred.User.Username)
	}
	if cred.PasswordHash != "hash-AdminUser" {
		t.Fatalf("password hash not returned for credential lookup")
	}

	byID, err := repo.GetUserByID(ctx, cred.User.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error: %v", err)
	}
	if byID.Username != "AdminUser" {
		t.Fatalf("GetUserByID username = %q", byID.Username)
	}
}

func TestUserRepositoryGetByUsernameNotFound(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	_, err := repo.GetUserByUsername(ctx, "nobody")
	if !errors.Is(err, appidentity.ErrNotFound) {
		t.Fatalf("GetUserByUsername() error = %v, want ErrNotFound", err)
	}
}

func TestUserRepositoryUpdatePasswordHashAndLastLogin(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	createTestUser(t, repo, "user-1", "alice")

	if err := repo.UpdateUserPasswordHash(ctx, domainidentity.UserID("user-1"), "new-hash"); err != nil {
		t.Fatalf("UpdateUserPasswordHash() error: %v", err)
	}
	cred, err := repo.GetUserByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetUserByUsername() error: %v", err)
	}
	if cred.PasswordHash != "new-hash" {
		t.Fatalf("password hash = %q, want new-hash", cred.PasswordHash)
	}

	lastLogin := "2026-07-12T12:00:00Z"
	if err := repo.UpdateUserLastLoginAt(ctx, domainidentity.UserID("user-1"), lastLogin); err != nil {
		t.Fatalf("UpdateUserLastLoginAt() error: %v", err)
	}
	got, err := repo.GetUserByID(ctx, domainidentity.UserID("user-1"))
	if err != nil {
		t.Fatalf("GetUserByID() error: %v", err)
	}
	if got.LastLoginAt != lastLogin {
		t.Fatalf("last_login_at = %q, want %q", got.LastLoginAt, lastLogin)
	}
}

func TestUserRepositorySetUserStatus(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	createTestUser(t, repo, "user-1", "alice")

	if err := repo.SetUserStatus(ctx, domainidentity.UserID("user-1"), appidentity.UserStatusDisabled); err != nil {
		t.Fatalf("SetUserStatus() error: %v", err)
	}
	got, err := repo.GetUserByID(ctx, domainidentity.UserID("user-1"))
	if err != nil {
		t.Fatalf("GetUserByID() error: %v", err)
	}
	if got.Status != appidentity.UserStatusDisabled {
		t.Fatalf("status = %q, want disabled", got.Status)
	}
}

func TestUserRepositoryListUsersBoundedAndDeterministic(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	for _, username := range []string{"charlie", "alice", "bob"} {
		createTestUser(t, repo, "user-"+username, username)
	}

	page1, err := repo.ListUsers(ctx, appidentity.ListUsersInput{Limit: 2})
	if err != nil {
		t.Fatalf("ListUsers page1 error: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}
	if page1[0].Username != "alice" || page1[1].Username != "bob" {
		t.Fatalf("page1 order = [%q, %q], want [alice, bob]", page1[0].Username, page1[1].Username)
	}

	page2, err := repo.ListUsers(ctx, appidentity.ListUsersInput{
		Limit:         2,
		AfterUsername: page1[len(page1)-1].Username,
	})
	if err != nil {
		t.Fatalf("ListUsers page2 error: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("page2 len = %d, want 1", len(page2))
	}
	if page2[0].Username != "charlie" {
		t.Fatalf("page2 user = %q, want charlie", page2[0].Username)
	}
}

func TestUserRepositoryCountUsers(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	count, err := repo.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers() error: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}

	createTestUser(t, repo, "user-1", "alice")
	createTestUser(t, repo, "user-2", "bob")

	count, err = repo.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers() error: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestUserRepositoryRejectsInvalidInput(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	err := repo.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           domainidentity.UserID("user-1"),
		Username:     "alice",
		PasswordHash: "hash",
		Role:         domainidentity.Role("invalid"),
		Status:       appidentity.UserStatusActive,
	})
	if !errors.Is(err, domainidentity.ErrInvalidRole) {
		t.Fatalf("CreateUser invalid role error = %v", err)
	}

	err = repo.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           domainidentity.UserID("user-1"),
		Username:     "alice",
		PasswordHash: "hash",
		Role:         domainidentity.RoleAdministrator,
		Status:       appidentity.UserStatus("locked"),
	})
	if err == nil || !strings.Contains(err.Error(), "invalid user status") {
		t.Fatalf("CreateUser invalid status error = %v", err)
	}

	_, err = repo.ListUsers(ctx, appidentity.ListUsersInput{Limit: 0})
	if err == nil || !strings.Contains(err.Error(), "limit must be positive") {
		t.Fatalf("ListUsers zero limit error = %v", err)
	}

	_, err = repo.ListUsers(ctx, appidentity.ListUsersInput{Limit: 101})
	if err == nil || !strings.Contains(err.Error(), "limit exceeds maximum") {
		t.Fatalf("ListUsers over max limit error = %v", err)
	}

	err = repo.SetUserStatus(ctx, domainidentity.UserID("user-1"), appidentity.UserStatus("locked"))
	if err == nil || !strings.Contains(err.Error(), "invalid user status") {
		t.Fatalf("SetUserStatus invalid status error = %v", err)
	}
}

func TestUserRepositoryDuplicateUsernameFailsSafely(t *testing.T) {
	store, ctx := openIdentityTestDB(t)
	repo := sqlite.NewUserRepository(store.DB())

	createTestUser(t, repo, "user-1", "alice")

	err := repo.CreateUser(ctx, appidentity.CreateUserInput{
		ID:           domainidentity.UserID("user-2"),
		Username:     "ALICE",
		PasswordHash: "other-hash",
		Role:         domainidentity.RoleUser,
		Status:       appidentity.UserStatusActive,
	})
	if err == nil {
		t.Fatal("CreateUser duplicate username expected error")
	}
	if strings.Contains(err.Error(), "other-hash") {
		t.Fatalf("duplicate username error leaked secret: %v", err)
	}
}

func TestUserStoreHasNoDeleteMethod(t *testing.T) {
	var _ appidentity.UserStore = (*sqlite.UserRepository)(nil)
}
