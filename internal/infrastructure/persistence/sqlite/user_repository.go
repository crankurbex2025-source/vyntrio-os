package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// UserRepository implements appidentity.UserStore.
type UserRepository struct {
	q *sqlcgen.Queries
}

var _ appidentity.UserStore = (*UserRepository)(nil)

// NewUserRepository creates a user store backed by the given database.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{q: sqlcgen.New(db)}
}

// CreateUser inserts a new user record.
func (r *UserRepository) CreateUser(ctx context.Context, input appidentity.CreateUserInput) error {
	if !input.Role.Valid() {
		return domainidentity.ErrInvalidRole
	}
	if input.Status != appidentity.UserStatusActive && input.Status != appidentity.UserStatusDisabled {
		return fmt.Errorf("invalid user status")
	}

	if err := r.q.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:                 string(input.ID),
		Username:           input.Username,
		DisplayName:        nullString(input.DisplayName),
		PasswordHash:       input.PasswordHash,
		Role:               input.Role.String(),
		Status:             string(input.Status),
		MustChangePassword: intFromBool(input.MustChangePassword),
	}); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUserByID returns a user without credential material.
func (r *UserRepository) GetUserByID(ctx context.Context, id domainidentity.UserID) (appidentity.User, error) {
	row, err := r.q.GetUserByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appidentity.User{}, appidentity.ErrNotFound
		}
		return appidentity.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return mapUserRow(row.ID, row.Username, row.DisplayName, row.Role, row.Status, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.LastLoginAt)
}

// GetUserByUsername returns credential lookup data including password hash.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (appidentity.UserCredential, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appidentity.UserCredential{}, appidentity.ErrNotFound
		}
		return appidentity.UserCredential{}, fmt.Errorf("get user by username: %w", err)
	}

	user, err := mapUserRow(row.ID, row.Username, row.DisplayName, row.Role, row.Status, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.LastLoginAt)
	if err != nil {
		return appidentity.UserCredential{}, err
	}
	return appidentity.UserCredential{
		User:         user,
		PasswordHash: row.PasswordHash,
	}, nil
}

// UpdateUserPasswordHash replaces the stored password hash.
func (r *UserRepository) UpdateUserPasswordHash(ctx context.Context, id domainidentity.UserID, passwordHash string) error {
	if err := r.q.UpdateUserPasswordHash(ctx, sqlcgen.UpdateUserPasswordHashParams{
		PasswordHash: passwordHash,
		ID:           string(id),
	}); err != nil {
		return fmt.Errorf("update user password hash: %w", err)
	}
	return nil
}

// UpdateUserLastLoginAt sets the last login timestamp.
func (r *UserRepository) UpdateUserLastLoginAt(ctx context.Context, id domainidentity.UserID, lastLoginAt string) error {
	if err := r.q.UpdateUserLastLoginAt(ctx, sqlcgen.UpdateUserLastLoginAtParams{
		LastLoginAt: nullString(lastLoginAt),
		ID:          string(id),
	}); err != nil {
		return fmt.Errorf("update user last login: %w", err)
	}
	return nil
}

// SetUserStatus updates the user status.
func (r *UserRepository) SetUserStatus(ctx context.Context, id domainidentity.UserID, status appidentity.UserStatus) error {
	if status != appidentity.UserStatusActive && status != appidentity.UserStatusDisabled {
		return fmt.Errorf("invalid user status")
	}
	if err := r.q.SetUserStatus(ctx, sqlcgen.SetUserStatusParams{
		Status: string(status),
		ID:     string(id),
	}); err != nil {
		return fmt.Errorf("set user status: %w", err)
	}
	return nil
}

// ListUsers returns a bounded page ordered by username.
func (r *UserRepository) ListUsers(ctx context.Context, input appidentity.ListUsersInput) ([]appidentity.User, error) {
	limit, err := validateListLimit(input.Limit)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListUsers(ctx, sqlcgen.ListUsersParams{
		AfterUsername: input.AfterUsername,
		RowLimit:      limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	out := make([]appidentity.User, 0, len(rows))
	for _, row := range rows {
		user, err := mapUserRow(row.ID, row.Username, row.DisplayName, row.Role, row.Status, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.LastLoginAt)
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		out = append(out, user)
	}
	return out, nil
}

// CountUsers returns the total number of users.
func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	count, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

func mapUserRow(
	id, username string,
	displayName sql.NullString,
	role, status string,
	mustChangePassword int64,
	createdAt, updatedAt string,
	lastLoginAt sql.NullString,
) (appidentity.User, error) {
	parsedRole, err := parseRole(role)
	if err != nil {
		return appidentity.User{}, err
	}
	return appidentity.User{
		ID:                 domainidentity.UserID(id),
		Username:           username,
		DisplayName:        stringFromNull(displayName),
		Role:               parsedRole,
		Status:             appidentity.UserStatus(status),
		MustChangePassword: boolFromInt(mustChangePassword),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		LastLoginAt:        stringFromNull(lastLoginAt),
	}, nil
}
