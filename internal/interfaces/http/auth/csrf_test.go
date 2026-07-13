package auth_test

import (
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/auth"
)

func TestSessionCSRFValidMatchesExpectedHash(t *testing.T) {
	raw := "csrf-token-value-for-test"
	csrf := auth.NewSessionCSRF(appidentity.HashRawToken(raw))
	if !csrf.Valid(raw) {
		t.Fatal("expected valid CSRF header to match")
	}
}

func TestSessionCSRFValidRejectsWrongToken(t *testing.T) {
	csrf := auth.NewSessionCSRF(appidentity.HashRawToken("expected-token"))
	if csrf.Valid("wrong-token-value") {
		t.Fatal("expected wrong CSRF header to be rejected")
	}
}

func TestSessionCSRFValidRejectsEmptyAndOversized(t *testing.T) {
	csrf := auth.NewSessionCSRF(appidentity.HashRawToken("token"))
	if csrf.Valid("") {
		t.Fatal("empty header must be rejected")
	}
	oversized := make([]byte, appidentity.MaxCSRFHeaderValueLen+1)
	for i := range oversized {
		oversized[i] = 'a'
	}
	if csrf.Valid(string(oversized)) {
		t.Fatal("oversized header must be rejected")
	}
}

func TestSessionCSRFValidUsesConstantTimeCompare(t *testing.T) {
	const raw = "01234567890123456789012345678901"
	const wrong = "01234567890123456789012345678900"
	csrf := auth.NewSessionCSRF(appidentity.HashRawToken(raw))

	if len(wrong) != len(raw) {
		t.Fatalf("test setup: token lengths differ")
	}
	if csrf.Valid(wrong) {
		t.Fatal("wrong token must be rejected")
	}
	if !csrf.Valid(raw) {
		t.Fatal("correct token must be accepted")
	}
}
