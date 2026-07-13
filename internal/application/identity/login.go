package identity

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrAuthenticationFailed indicates credentials could not be verified.
var ErrAuthenticationFailed = errors.New("authentication failed")

// ErrInvalidLoginPassword indicates the login password is invalid for request validation.
var ErrInvalidLoginPassword = errors.New("invalid login password")

// LoginSessionCreator atomically persists a session and login audit event.
type LoginSessionCreator interface {
	CreateSessionWithAudit(ctx context.Context, session CreateSessionInput, audit AppendSecurityAuditEventInput) error
}

// LoginService authenticates users and creates server-side sessions.
type LoginService struct {
	users   UserStore
	hasher  credentialHasher
	tokens  *SessionTokenService
	creator LoginSessionCreator
	audit   SecurityAuditStore
	now     func() time.Time
}

type credentialHasher interface {
	HashPassword(ctx context.Context, plaintext string) (string, error)
	VerifyPassword(ctx context.Context, plaintext, encodedHash string) (valid bool, needsRehash bool, err error)
}

// NewLoginService returns a login service.
func NewLoginService(
	users UserStore,
	hasher credentialHasher,
	tokens *SessionTokenService,
	creator LoginSessionCreator,
	audit SecurityAuditStore,
) *LoginService {
	return &LoginService{
		users:   users,
		hasher:  hasher,
		tokens:  tokens,
		creator: creator,
		audit:   audit,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

// LoginResult holds session material for immediate cookie issuance only.
type LoginResult struct {
	Material SessionMaterial
}

// Login validates credentials, optionally rehashes the password, and creates a session.
func (s *LoginService) Login(
	ctx context.Context,
	username, password, sessionID, auditID string,
) (LoginResult, error) {
	if err := ctx.Err(); err != nil {
		return LoginResult{}, err
	}
	if err := ValidateLoginUsername(username); err != nil {
		return LoginResult{}, err
	}
	if err := validateLoginPassword(password); err != nil {
		return LoginResult{}, err
	}

	userCred, err := s.authenticateUser(ctx, username, password)
	if err != nil {
		if errors.Is(err, ErrAuthenticationFailed) {
			if auditErr := s.audit.AppendSecurityAuditEvent(ctx, AppendSecurityAuditEventInput{
				ID:           auditID,
				EventType:    AuditEventLoginFailure,
				Result:       "failure",
				MetadataJSON: `{}`,
			}); auditErr != nil {
				return LoginResult{}, auditErr
			}
		}
		return LoginResult{}, err
	}

	now := s.now()
	material, err := s.tokens.NewSessionMaterial(ctx, now)
	if err != nil {
		return LoginResult{}, err
	}

	if err := s.creator.CreateSessionWithAudit(ctx, CreateSessionInput{
		ID:               sessionID,
		UserID:           userCred.User.ID,
		SessionTokenHash: material.SessionTokenHash,
		CSRFTokenHash:    material.CSRFTokenHash,
		CreatedAt:        FormatUTCTime(material.CreatedAt),
		LastSeenAt:       FormatUTCTime(material.CreatedAt),
		ExpiresAt:        FormatUTCTime(material.ExpiresAt),
		IdleExpiresAt:    FormatUTCTime(material.IdleExpiresAt),
	}, AppendSecurityAuditEventInput{
		ID:            auditID,
		ActorUserID:   userCred.User.ID,
		SubjectUserID: userCred.User.ID,
		EventType:     AuditEventLoginSucceeded,
		Result:        "success",
		MetadataJSON:  `{}`,
	}); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Material: material}, nil
}

func (s *LoginService) authenticateUser(ctx context.Context, username, password string) (UserCredential, error) {
	userCred, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return UserCredential{}, ErrAuthenticationFailed
		}
		return UserCredential{}, err
	}
	if userCred.User.Status != UserStatusActive {
		return UserCredential{}, ErrAuthenticationFailed
	}

	valid, needsRehash, err := s.hasher.VerifyPassword(ctx, password, userCred.PasswordHash)
	if err != nil {
		return UserCredential{}, err
	}
	if !valid {
		return UserCredential{}, ErrAuthenticationFailed
	}

	if needsRehash {
		newHash, err := s.hasher.HashPassword(ctx, password)
		if err != nil {
			return UserCredential{}, err
		}
		if err := s.users.UpdateUserPasswordHash(ctx, userCred.User.ID, newHash); err != nil {
			return UserCredential{}, fmt.Errorf("update password hash: %w", err)
		}
	}

	return userCred, nil
}

func validateLoginPassword(password string) error {
	if password == "" || len(password) > maxPasswordBytes {
		return ErrInvalidLoginPassword
	}
	return nil
}

// ValidateLoginUsername applies canonical username rules for authentication.
func ValidateLoginUsername(username string) error {
	return ValidateBootstrapUsername(username)
}

// IsLoginClientInputError reports whether the error is a safe client input validation failure.
func IsLoginClientInputError(err error) bool {
	return errors.Is(err, ErrInvalidBootstrapUsername) ||
		errors.Is(err, ErrInvalidLoginPassword)
}

// IsLoginAuthenticationError reports whether the error is a generic credential rejection.
func IsLoginAuthenticationError(err error) bool {
	return errors.Is(err, ErrAuthenticationFailed)
}
