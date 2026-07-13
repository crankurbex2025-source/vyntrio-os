package identity

import (
	"context"
	"errors"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

const minBootstrapPasswordBytes = 12

// ErrBootstrapUnavailable indicates bootstrap cannot create a first user.
var ErrBootstrapUnavailable = errors.New("bootstrap unavailable")

// ErrInvalidBootstrapPassword indicates the bootstrap password is invalid.
var ErrInvalidBootstrapPassword = errors.New("invalid bootstrap password")

// BootstrapCreateInput holds server-derived first-owner creation data.
type BootstrapCreateInput struct {
	UserID             string
	Username           string
	PasswordHash       string
	Role               string
	Status             string
	MustChangePassword bool
}

// BootstrapAuditInput holds server-derived bootstrap audit data.
type BootstrapAuditInput struct {
	ID            string
	ActorUserID   string
	SubjectUserID string
	EventType     string
	Result        string
	MetadataJSON  string
}

// BootstrapCreator atomically creates the first user and audit event.
type BootstrapCreator interface {
	CreateFirstOwner(ctx context.Context, user BootstrapCreateInput, audit BootstrapAuditInput) (created bool, err error)
}

// BootstrapService creates the initial owner account during bootstrap.
type BootstrapService struct {
	hasher  passwordHasher
	creator BootstrapCreator
	users   UserStore
}

type passwordHasher interface {
	HashPassword(ctx context.Context, plaintext string) (string, error)
}

// NewBootstrapService returns a bootstrap service.
func NewBootstrapService(hasher passwordHasher, creator BootstrapCreator, users UserStore) *BootstrapService {
	return &BootstrapService{hasher: hasher, creator: creator, users: users}
}

// BootstrapResult is the successful first-owner creation result.
type BootstrapResult struct {
	UserID   string
	Username string
	Role     string
	Status   string
}

// CreateFirstOwner validates input, hashes the password and creates the owner if none exist.
func (s *BootstrapService) CreateFirstOwner(ctx context.Context, username, password, userID, auditID string) (BootstrapResult, error) {
	if err := ctx.Err(); err != nil {
		return BootstrapResult{}, err
	}
	if err := ValidateBootstrapUsername(username); err != nil {
		return BootstrapResult{}, err
	}
	if err := validateBootstrapPassword(password); err != nil {
		return BootstrapResult{}, err
	}

	count, err := s.users.CountUsers(ctx)
	if err != nil {
		return BootstrapResult{}, err
	}
	if count != 0 {
		return BootstrapResult{}, ErrBootstrapUnavailable
	}

	passwordHash, err := s.hasher.HashPassword(ctx, password)
	if err != nil {
		return BootstrapResult{}, err
	}

	created, err := s.creator.CreateFirstOwner(ctx, BootstrapCreateInput{
		UserID:             userID,
		Username:           username,
		PasswordHash:       passwordHash,
		Role:               domainidentity.RoleOwner.String(),
		Status:             string(UserStatusActive),
		MustChangePassword: false,
	}, BootstrapAuditInput{
		ID:            auditID,
		ActorUserID:   userID,
		SubjectUserID: userID,
		EventType:     "identity.bootstrap.succeeded",
		Result:        "success",
		MetadataJSON:  `{"source":"loopback"}`,
	})
	if err != nil {
		return BootstrapResult{}, err
	}
	if !created {
		return BootstrapResult{}, ErrBootstrapUnavailable
	}

	return BootstrapResult{
		UserID:   userID,
		Username: username,
		Role:     domainidentity.RoleOwner.String(),
		Status:   string(UserStatusActive),
	}, nil
}

func validateBootstrapPassword(password string) error {
	if password == "" || len(password) < minBootstrapPasswordBytes || len(password) > maxPasswordBytes {
		return ErrInvalidBootstrapPassword
	}
	return nil
}

func IsBootstrapClientInputError(err error) bool {
	return errors.Is(err, ErrInvalidBootstrapUsername) ||
		errors.Is(err, ErrInvalidBootstrapPassword) ||
		errors.Is(err, ErrEmptyPassword) ||
		errors.Is(err, ErrPasswordTooLong)
}

func IsBootstrapUnavailableError(err error) bool {
	return errors.Is(err, ErrBootstrapUnavailable)
}
