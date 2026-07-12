package identity_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
)

func newTestSessionTokenService(t *testing.T, config appidentity.SessionTokenConfig) *appidentity.SessionTokenService {
	t.Helper()

	svc, err := appidentity.NewSessionTokenService(config)
	if err != nil {
		t.Fatalf("NewSessionTokenService() error: %v", err)
	}
	return svc
}

func TestNewSessionTokenServiceAcceptsDefaults(t *testing.T) {
	if _, err := appidentity.NewSessionTokenService(appidentity.DefaultSessionTokenConfig); err != nil {
		t.Fatalf("NewSessionTokenService(default) error: %v", err)
	}
}

func TestNewSessionTokenServiceRejectsInvalidConfig(t *testing.T) {
	base := appidentity.DefaultSessionTokenConfig

	cases := []appidentity.SessionTokenConfig{
		func() appidentity.SessionTokenConfig { c := base; c.AbsoluteTTL = 0; return c }(),
		func() appidentity.SessionTokenConfig { c := base; c.IdleTTL = 0; return c }(),
		func() appidentity.SessionTokenConfig {
			return appidentity.SessionTokenConfig{
				AbsoluteTTL:       time.Hour,
				IdleTTL:           2 * time.Hour,
				SessionTokenBytes: 32,
				CSRFTokenBytes:    32,
			}
		}(),
		func() appidentity.SessionTokenConfig { c := base; c.SessionTokenBytes = 16; return c }(),
		func() appidentity.SessionTokenConfig { c := base; c.CSRFTokenBytes = 16; return c }(),
		func() appidentity.SessionTokenConfig { c := base; c.SessionTokenBytes = 128; return c }(),
		func() appidentity.SessionTokenConfig { c := base; c.CSRFTokenBytes = 128; return c }(),
	}

	for i, config := range cases {
		if _, err := appidentity.NewSessionTokenService(config); !errors.Is(err, appidentity.ErrInvalidSessionTokenConfig) {
			t.Fatalf("case %d: expected ErrInvalidSessionTokenConfig, got %v", i, err)
		}
	}
}

func TestNewSessionMaterialLifetimeFromSuppliedUTC(t *testing.T) {
	config := appidentity.SessionTokenConfig{
		AbsoluteTTL:       2 * time.Hour,
		IdleTTL:           30 * time.Minute,
		SessionTokenBytes: 32,
		CSRFTokenBytes:    32,
	}
	svc := newTestSessionTokenService(t, config)

	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	material, err := svc.NewSessionMaterial(context.Background(), now)
	if err != nil {
		t.Fatalf("NewSessionMaterial() error: %v", err)
	}

	if !material.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt not equal to supplied now")
	}
	if !material.ExpiresAt.Equal(now.Add(2 * time.Hour)) {
		t.Fatalf("ExpiresAt mismatch")
	}
	if !material.IdleExpiresAt.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("IdleExpiresAt mismatch")
	}
	if material.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt location = %v, want UTC", material.CreatedAt.Location())
	}
}

func TestNewSessionMaterialNormalizesNonUTCTime(t *testing.T) {
	svc := newTestSessionTokenService(t, appidentity.DefaultSessionTokenConfig)

	loc := time.FixedZone("CEST", 2*3600)
	local := time.Date(2026, 7, 12, 12, 0, 0, 0, loc)
	expectedUTC := local.UTC()

	material, err := svc.NewSessionMaterial(context.Background(), local)
	if err != nil {
		t.Fatalf("NewSessionMaterial() error: %v", err)
	}
	if !material.CreatedAt.Equal(expectedUTC) {
		t.Fatalf("CreatedAt not normalized to UTC")
	}
}

func TestNewSessionMaterialGeneratesDistinctValues(t *testing.T) {
	svc := newTestSessionTokenService(t, appidentity.DefaultSessionTokenConfig)
	ctx := context.Background()
	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)

	first, err := svc.NewSessionMaterial(ctx, now)
	if err != nil {
		t.Fatalf("first NewSessionMaterial() error: %v", err)
	}
	second, err := svc.NewSessionMaterial(ctx, now)
	if err != nil {
		t.Fatalf("second NewSessionMaterial() error: %v", err)
	}

	if first.RawSessionToken == second.RawSessionToken {
		t.Fatal("expected distinct raw session tokens")
	}
	if first.RawCSRFToken == second.RawCSRFToken {
		t.Fatal("expected distinct raw csrf tokens")
	}
	if first.SessionTokenHash == second.SessionTokenHash {
		t.Fatal("expected distinct session token hashes")
	}
	if first.CSRFTokenHash == second.CSRFTokenHash {
		t.Fatal("expected distinct csrf token hashes")
	}
}

func TestNewSessionMaterialSessionAndCSRFTokensDiffer(t *testing.T) {
	svc := newTestSessionTokenService(t, appidentity.DefaultSessionTokenConfig)

	material, err := svc.NewSessionMaterial(context.Background(), time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewSessionMaterial() error: %v", err)
	}
	if material.RawSessionToken == material.RawCSRFToken {
		t.Fatal("raw session and csrf tokens must differ")
	}
	if material.SessionTokenHash == material.CSRFTokenHash {
		t.Fatal("session and csrf token hashes must differ")
	}
}

func TestNewSessionMaterialRawTokenEncoding(t *testing.T) {
	config := appidentity.DefaultSessionTokenConfig
	svc := newTestSessionTokenService(t, config)

	material, err := svc.NewSessionMaterial(context.Background(), time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewSessionMaterial() error: %v", err)
	}

	if strings.Contains(material.RawSessionToken, "=") || strings.Contains(material.RawCSRFToken, "=") {
		t.Fatal("raw token must not contain base64 padding")
	}

	sessionDecoded, err := base64.RawURLEncoding.DecodeString(material.RawSessionToken)
	if err != nil {
		t.Fatalf("session token decode error: %v", err)
	}
	if len(sessionDecoded) != config.SessionTokenBytes {
		t.Fatalf("session token length = %d, want %d", len(sessionDecoded), config.SessionTokenBytes)
	}

	csrfDecoded, err := base64.RawURLEncoding.DecodeString(material.RawCSRFToken)
	if err != nil {
		t.Fatalf("csrf token decode error: %v", err)
	}
	if len(csrfDecoded) != config.CSRFTokenBytes {
		t.Fatalf("csrf token length = %d, want %d", len(csrfDecoded), config.CSRFTokenBytes)
	}
}

func TestNewSessionMaterialStorageHashes(t *testing.T) {
	svc := newTestSessionTokenService(t, appidentity.DefaultSessionTokenConfig)

	material, err := svc.NewSessionMaterial(context.Background(), time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewSessionMaterial() error: %v", err)
	}

	assertStorageHash(t, material.RawSessionToken, material.SessionTokenHash)
	assertStorageHash(t, material.RawCSRFToken, material.CSRFTokenHash)
}

func assertStorageHash(t *testing.T, rawToken, storedHash string) {
	t.Helper()

	if len(storedHash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(storedHash))
	}
	if storedHash != strings.ToLower(storedHash) {
		t.Fatal("hash must be lowercase hex")
	}
	if _, err := hex.DecodeString(storedHash); err != nil {
		t.Fatalf("hash is not valid hex: %v", err)
	}

	sum := sha256.Sum256([]byte(rawToken))
	expected := hex.EncodeToString(sum[:])
	if storedHash != expected {
		t.Fatal("stored hash does not match SHA-256 of raw token")
	}
}

func TestNewSessionMaterialCancelledContext(t *testing.T) {
	svc := newTestSessionTokenService(t, appidentity.DefaultSessionTokenConfig)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.NewSessionMaterial(ctx, time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context error = %v, want context.Canceled", err)
	}
}

func TestNewSessionMaterialRejectsZeroTime(t *testing.T) {
	svc := newTestSessionTokenService(t, appidentity.DefaultSessionTokenConfig)

	_, err := svc.NewSessionMaterial(context.Background(), time.Time{})
	if !errors.Is(err, appidentity.ErrInvalidSessionTime) {
		t.Fatalf("zero time error = %v, want ErrInvalidSessionTime", err)
	}
}

func TestDefaultSessionTokenConfigMatchesADR(t *testing.T) {
	cfg := appidentity.DefaultSessionTokenConfig
	if cfg.AbsoluteTTL != 7*24*time.Hour {
		t.Fatalf("AbsoluteTTL = %v, want 7 days", cfg.AbsoluteTTL)
	}
	if cfg.IdleTTL != 24*time.Hour {
		t.Fatalf("IdleTTL = %v, want 24 hours", cfg.IdleTTL)
	}
}
