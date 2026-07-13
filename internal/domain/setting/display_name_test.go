package setting_test

import (
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

func TestValidateInstanceDisplayNameAcceptsTrimmedName(t *testing.T) {
	got, err := setting.ValidateInstanceDisplayName("  Vyntrio Home  ")
	if err != nil {
		t.Fatalf("ValidateInstanceDisplayName() error: %v", err)
	}
	if got != "Vyntrio Home" {
		t.Fatalf("got = %q", got)
	}
}

func TestValidateInstanceDisplayNameRejectsControlCharacters(t *testing.T) {
	if _, err := setting.ValidateInstanceDisplayName("bad\nname"); err == nil {
		t.Fatal("expected control character to fail")
	}
}

func TestValidateInstanceDisplayNameRejectsOverlongName(t *testing.T) {
	if _, err := setting.ValidateInstanceDisplayName(strings.Repeat("a", setting.MaxInstanceDisplayNameRunes+1)); err == nil {
		t.Fatal("expected overlong name to fail")
	}
}
