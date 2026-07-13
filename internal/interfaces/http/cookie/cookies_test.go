package cookie_test

import (
	"net/http/httptest"
	"testing"
	"time"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/interfaces/http/cookie"
)

func TestNewPolicyProductionSecureDefault(t *testing.T) {
	policy := cookie.NewPolicy("production", nil)
	if !policy.Secure {
		t.Fatal("production must default to Secure cookies")
	}
}

func TestNewPolicyDevelopmentInsecureDefault(t *testing.T) {
	policy := cookie.NewPolicy("development", nil)
	if policy.Secure {
		t.Fatal("development must default to insecure cookies")
	}
}

func TestNewPolicyExplicitOverride(t *testing.T) {
	secure := true
	policy := cookie.NewPolicy("development", &secure)
	if !policy.Secure {
		t.Fatal("explicit override must enable Secure cookies")
	}
}

func TestSetAndClearSessionCookie(t *testing.T) {
	secure := true
	policy := cookie.NewPolicy("production", &secure)
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	expires := now.Add(7 * 24 * time.Hour)
	material := appidentity.SessionMaterial{
		RawSessionToken: "session-token",
		ExpiresAt:       expires,
	}

	rec := httptest.NewRecorder()
	policy.SetSessionCookie(rec, material, now)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("set cookie count = %d, want 1", len(cookies))
	}
	if cookies[0].Name != cookie.SessionCookieName {
		t.Fatalf("cookie name = %q", cookies[0].Name)
	}

	clearRec := httptest.NewRecorder()
	policy.ClearSessionCookie(clearRec)
	if len(clearRec.Result().Cookies()) != 1 {
		t.Fatalf("clear cookie count = %d, want 1", len(clearRec.Result().Cookies()))
	}
}
