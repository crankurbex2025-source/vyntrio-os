package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	minSessionTokenBytes = 32
	maxSessionTokenBytes = 64
)

// ErrInvalidSessionTokenConfig indicates session token configuration is invalid.
var ErrInvalidSessionTokenConfig = errors.New("invalid session token configuration")

// ErrInvalidSessionTime indicates the supplied session timestamp is invalid.
var ErrInvalidSessionTime = errors.New("invalid session time")

// SessionTokenConfig holds validated session and CSRF token generation settings.
type SessionTokenConfig struct {
	AbsoluteTTL       time.Duration
	IdleTTL           time.Duration
	SessionTokenBytes int
	CSRFTokenBytes    int
}

// DefaultSessionTokenConfig is aligned with ADR-0004 session TTL defaults.
var DefaultSessionTokenConfig = SessionTokenConfig{
	AbsoluteTTL:       7 * 24 * time.Hour,
	IdleTTL:           24 * time.Hour,
	SessionTokenBytes: 32,
	CSRFTokenBytes:    32,
}

// Validate checks session token configuration bounds.
func (c SessionTokenConfig) Validate() error {
	if c.AbsoluteTTL <= 0 {
		return fmt.Errorf("%w: absolute TTL must be positive", ErrInvalidSessionTokenConfig)
	}
	if c.IdleTTL <= 0 {
		return fmt.Errorf("%w: idle TTL must be positive", ErrInvalidSessionTokenConfig)
	}
	if c.IdleTTL > c.AbsoluteTTL {
		return fmt.Errorf("%w: idle TTL must not exceed absolute TTL", ErrInvalidSessionTokenConfig)
	}
	if c.SessionTokenBytes < minSessionTokenBytes || c.SessionTokenBytes > maxSessionTokenBytes {
		return fmt.Errorf("%w: session token bytes must be between %d and %d", ErrInvalidSessionTokenConfig, minSessionTokenBytes, maxSessionTokenBytes)
	}
	if c.CSRFTokenBytes < minSessionTokenBytes || c.CSRFTokenBytes > maxSessionTokenBytes {
		return fmt.Errorf("%w: csrf token bytes must be between %d and %d", ErrInvalidSessionTokenConfig, minSessionTokenBytes, maxSessionTokenBytes)
	}
	return nil
}

// SessionMaterial holds freshly generated session credentials for immediate caller use.
type SessionMaterial struct {
	RawSessionToken  string
	RawCSRFToken     string
	SessionTokenHash string
	CSRFTokenHash    string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	IdleExpiresAt    time.Time
}

// SessionTokenService generates opaque session and CSRF token material.
type SessionTokenService struct {
	config SessionTokenConfig
}

// NewSessionTokenService validates config and returns a session token service.
func NewSessionTokenService(config SessionTokenConfig) (*SessionTokenService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &SessionTokenService{config: config}, nil
}

// NewSessionMaterial generates raw tokens, storage hashes and lifetime timestamps.
func (s *SessionTokenService) NewSessionMaterial(ctx context.Context, now time.Time) (SessionMaterial, error) {
	if err := ctx.Err(); err != nil {
		return SessionMaterial{}, err
	}
	if now.IsZero() {
		return SessionMaterial{}, ErrInvalidSessionTime
	}

	createdAt := now.UTC()
	expiresAt := createdAt.Add(s.config.AbsoluteTTL)
	idleExpiresAt := createdAt.Add(s.config.IdleTTL)

	rawSessionToken, err := generateRawToken(ctx, s.config.SessionTokenBytes)
	if err != nil {
		return SessionMaterial{}, err
	}

	rawCSRFToken, err := generateRawToken(ctx, s.config.CSRFTokenBytes)
	if err != nil {
		return SessionMaterial{}, err
	}

	if rawSessionToken == rawCSRFToken {
		rawCSRFToken, err = generateRawToken(ctx, s.config.CSRFTokenBytes)
		if err != nil {
			return SessionMaterial{}, err
		}
		if rawSessionToken == rawCSRFToken {
			return SessionMaterial{}, fmt.Errorf("%w: session and csrf tokens must differ", ErrInvalidSessionTokenConfig)
		}
	}

	return SessionMaterial{
		RawSessionToken:  rawSessionToken,
		RawCSRFToken:     rawCSRFToken,
		SessionTokenHash: hashRawToken(rawSessionToken),
		CSRFTokenHash:    hashRawToken(rawCSRFToken),
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
		IdleExpiresAt:    idleExpiresAt,
	}, nil
}

func generateRawToken(ctx context.Context, size int) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashRawToken returns the SHA-256 hex digest used for session and CSRF storage lookup.
func HashRawToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func hashRawToken(rawToken string) string {
	return HashRawToken(rawToken)
}
