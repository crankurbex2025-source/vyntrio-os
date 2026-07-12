package identity_test

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
)

var testArgon2Config = appidentity.Argon2idConfig{
	Memory:      4096,
	Iterations:  1,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

var weakArgon2Config = appidentity.Argon2idConfig{
	Memory:      2048,
	Iterations:  1,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

func newTestHasher(t *testing.T, config appidentity.Argon2idConfig) *appidentity.PasswordHasher {
	t.Helper()

	hasher, err := appidentity.NewPasswordHasher(config)
	if err != nil {
		t.Fatalf("NewPasswordHasher() error: %v", err)
	}
	return hasher
}

func TestNewPasswordHasherRejectsInvalidConfig(t *testing.T) {
	cases := []appidentity.Argon2idConfig{
		{Memory: 0, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32},
		{Memory: 4096, Iterations: 0, Parallelism: 1, SaltLength: 16, KeyLength: 32},
		{Memory: 4096, Iterations: 1, Parallelism: 0, SaltLength: 16, KeyLength: 32},
		{Memory: 4096, Iterations: 1, Parallelism: 1, SaltLength: 4, KeyLength: 32},
		{Memory: 4096, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 8},
	}

	for i, config := range cases {
		if _, err := appidentity.NewPasswordHasher(config); !errors.Is(err, appidentity.ErrInvalidArgon2Config) {
			t.Fatalf("case %d: expected ErrInvalidArgon2Config, got %v", i, err)
		}
	}
}

func TestHashPasswordGeneratesDistinctHashes(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()

	first, err := hasher.HashPassword(ctx, "correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword first error: %v", err)
	}
	second, err := hasher.HashPassword(ctx, "correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword second error: %v", err)
	}
	if first == second {
		t.Fatal("expected different encoded hashes for same plaintext")
	}
}

func TestEncodedHashStructure(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	encoded, err := hasher.HashPassword(context.Background(), "structure-check-password")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		t.Fatalf("unexpected encoded structure")
	}
	if !strings.HasPrefix(parts[3], "m=") || !strings.Contains(parts[3], ",t=") || !strings.Contains(parts[3], ",p=") {
		t.Fatalf("unexpected params segment: %q", parts[3])
	}
	if _, err := base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
		t.Fatalf("salt base64 decode error: %v", err)
	}
	if _, err := base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
		t.Fatalf("hash base64 decode error: %v", err)
	}
}

func TestVerifyPasswordValidSameConfig(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()

	encoded, err := hasher.HashPassword(ctx, "verify-me-please")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	valid, needsRehash, err := hasher.VerifyPassword(ctx, "verify-me-please", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if !valid || needsRehash {
		t.Fatalf("valid=%v needsRehash=%v, want true/false", valid, needsRehash)
	}
}

func TestVerifyPasswordWrongPassword(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()

	encoded, err := hasher.HashPassword(ctx, "actual-password")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	valid, needsRehash, err := hasher.VerifyPassword(ctx, "wrong-password", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if valid || needsRehash {
		t.Fatalf("valid=%v needsRehash=%v, want false/false", valid, needsRehash)
	}
}

func TestHashAndVerifyRejectEmptyPassword(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()

	if _, err := hasher.HashPassword(ctx, ""); !errors.Is(err, appidentity.ErrEmptyPassword) {
		t.Fatalf("HashPassword empty error = %v", err)
	}

	valid, needsRehash, err := hasher.VerifyPassword(ctx, "", "dummy")
	if !errors.Is(err, appidentity.ErrEmptyPassword) {
		t.Fatalf("VerifyPassword empty error = %v", err)
	}
	if valid || needsRehash {
		t.Fatalf("valid=%v needsRehash=%v, want false/false", valid, needsRehash)
	}
}

func TestVerifyPasswordRejectsNonCanonicalParameterSegment(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()
	plaintext := "canonical-params-test"

	validHash, err := hasher.HashPassword(ctx, plaintext)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	valid, needsRehash, err := hasher.VerifyPassword(ctx, plaintext, validHash)
	if err != nil {
		t.Fatalf("VerifyPassword() canonical hash error: %v", err)
	}
	if !valid || needsRehash {
		t.Fatalf("canonical hash valid=%v needsRehash=%v, want true/false", valid, needsRehash)
	}

	parts := strings.Split(validHash, "$")
	validSalt := parts[4]
	validKey := parts[5]

	nonCanonical := []string{
		"$argon2id$v=19$t=1,m=4096,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,m=8192,t=1,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,t=2,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1,p=2$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096, t=1,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1 $" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1,q=1$" + validSalt + "$" + validKey,
	}

	for i, encoded := range nonCanonical {
		valid, needsRehash, err := hasher.VerifyPassword(ctx, plaintext, encoded)
		if err != nil {
			t.Fatalf("case %d: VerifyPassword() error = %v", i, err)
		}
		if valid || needsRehash {
			t.Fatalf("case %d: valid=%v needsRehash=%v, want false/false", i, valid, needsRehash)
		}
	}
}

func TestVerifyPasswordMalformedEncodings(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()
	plaintext := "password-for-malformed-tests"

	validHash, err := hasher.HashPassword(ctx, plaintext)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	parts := strings.Split(validHash, "$")
	validSalt := parts[4]
	validKey := parts[5]

	malformed := []string{
		"",
		"not-a-hash",
		"$argon2i$v=19$m=4096,t=1,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=18$m=4096,t=1,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=0,t=1,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=0,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=0$" + validSalt + "$" + validKey,
		"$argon2id$v=19x$m=4096,t=1,p=1$" + validSalt + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1$$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1$" + validSalt + "$",
		"$argon2id$v=19$m=4096,t=1,p=1$!!!$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1$" + validSalt + "$!!!",
		"$argon2id$v=19$m=4096,t=1,p=1$" + base64.RawStdEncoding.EncodeToString([]byte("short")) + "$" + validKey,
		"$argon2id$v=19$m=4096,t=1,p=1$" + validSalt + "$" + base64.RawStdEncoding.EncodeToString([]byte("short")),
	}

	for i, encoded := range malformed {
		valid, needsRehash, err := hasher.VerifyPassword(ctx, plaintext, encoded)
		if err != nil {
			t.Fatalf("case %d: VerifyPassword() error = %v", i, err)
		}
		if valid || needsRehash {
			t.Fatalf("case %d: valid=%v needsRehash=%v, want false/false", i, valid, needsRehash)
		}
	}
}

func TestVerifyPasswordNeedsRehashForWeakerConfig(t *testing.T) {
	weakHasher := newTestHasher(t, weakArgon2Config)
	currentHasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()

	encoded, err := weakHasher.HashPassword(ctx, "rehash-me")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	valid, needsRehash, err := currentHasher.VerifyPassword(ctx, "rehash-me", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if !valid || !needsRehash {
		t.Fatalf("valid=%v needsRehash=%v, want true/true", valid, needsRehash)
	}
}

func TestPasswordMaximumByteLength(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)
	ctx := context.Background()

	withinLimit := strings.Repeat("a", 1024)
	if _, err := hasher.HashPassword(ctx, withinLimit); err != nil {
		t.Fatalf("HashPassword at limit error: %v", err)
	}

	overLimitASCII := strings.Repeat("a", 1025)
	if _, err := hasher.HashPassword(ctx, overLimitASCII); !errors.Is(err, appidentity.ErrPasswordTooLong) {
		t.Fatalf("HashPassword over ASCII limit error = %v", err)
	}

	unicodeOverLimit := strings.Repeat("é", 513) // 513 runes * 2 bytes = 1026 bytes
	if _, err := hasher.HashPassword(ctx, unicodeOverLimit); !errors.Is(err, appidentity.ErrPasswordTooLong) {
		t.Fatalf("HashPassword over unicode byte limit error = %v", err)
	}
}

func TestHashPasswordCancelledContext(t *testing.T) {
	hasher := newTestHasher(t, testArgon2Config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := hasher.HashPassword(ctx, "cancelled-before-work"); !errors.Is(err, context.Canceled) {
		t.Fatalf("HashPassword cancelled error = %v", err)
	}
}

func TestDefaultArgon2ConfigAccepted(t *testing.T) {
	if _, err := appidentity.NewPasswordHasher(appidentity.DefaultArgon2idConfig); err != nil {
		t.Fatalf("NewPasswordHasher(default) error: %v", err)
	}
}
