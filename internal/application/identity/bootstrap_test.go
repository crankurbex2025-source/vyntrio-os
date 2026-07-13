package identity_test

import (
	"context"
	"errors"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

type stubUserStore struct {
	count int64
	err   error
}

func (s stubUserStore) CreateUser(context.Context, appidentity.CreateUserInput) error {
	panic("not implemented")
}
func (s stubUserStore) GetUserByID(context.Context, domainidentity.UserID) (appidentity.User, error) {
	panic("not implemented")
}
func (s stubUserStore) GetUserByUsername(context.Context, string) (appidentity.UserCredential, error) {
	panic("not implemented")
}
func (s stubUserStore) UpdateUserPasswordHash(context.Context, domainidentity.UserID, string) error {
	panic("not implemented")
}
func (s stubUserStore) UpdateUserLastLoginAt(context.Context, domainidentity.UserID, string) error {
	panic("not implemented")
}
func (s stubUserStore) SetUserStatus(context.Context, domainidentity.UserID, appidentity.UserStatus) error {
	panic("not implemented")
}
func (s stubUserStore) ListUsers(context.Context, appidentity.ListUsersInput) ([]appidentity.User, error) {
	panic("not implemented")
}
func (s stubUserStore) CountUsers(context.Context) (int64, error) {
	return s.count, s.err
}

type spyHasher struct {
	calls int
}

func (s *spyHasher) HashPassword(context.Context, string) (string, error) {
	s.calls++
	return "encoded-hash", nil
}

type stubCreator struct {
	calls int
}

func (s *stubCreator) CreateFirstOwner(context.Context, appidentity.BootstrapCreateInput, appidentity.BootstrapAuditInput) (bool, error) {
	s.calls++
	return true, nil
}

func TestBootstrapServiceSkipsHashWhenUsersExist(t *testing.T) {
	hasher := &spyHasher{}
	creator := &stubCreator{}
	svc := appidentity.NewBootstrapService(hasher, creator, stubUserStore{count: 1})

	_, err := svc.CreateFirstOwner(context.Background(), "owner", "valid-password-123", "user-id", "audit-id")
	if !errors.Is(err, appidentity.ErrBootstrapUnavailable) {
		t.Fatalf("error = %v, want ErrBootstrapUnavailable", err)
	}
	if hasher.calls != 0 {
		t.Fatalf("hasher calls = %d, want 0", hasher.calls)
	}
	if creator.calls != 0 {
		t.Fatalf("creator calls = %d, want 0", creator.calls)
	}
}
