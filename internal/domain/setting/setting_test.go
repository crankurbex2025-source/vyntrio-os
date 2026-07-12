package setting_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
)

func TestValidateKeyAllowed(t *testing.T) {
	for _, key := range []string{setting.KeyTimezone, setting.KeyHostname} {
		if err := setting.ValidateKey(setting.NamespaceSystem, key); err != nil {
			t.Fatalf("ValidateKey(%q) error: %v", key, err)
		}
	}
}

func TestValidateKeyRejectsUnknownKey(t *testing.T) {
	err := setting.ValidateKey(setting.NamespaceSystem, "locale")
	if !errors.Is(err, setting.ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestValidateNamespaceRejectsNonSystem(t *testing.T) {
	err := setting.ValidateKey("network", setting.KeyTimezone)
	if !errors.Is(err, setting.ErrInvalidNamespace) {
		t.Fatalf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestValidateValueTimezone(t *testing.T) {
	if err := setting.ValidateValue(setting.KeyTimezone, "UTC"); err != nil {
		t.Fatalf("ValidateValue timezone: %v", err)
	}
	if err := setting.ValidateValue(setting.KeyTimezone, ""); err == nil {
		t.Fatal("expected empty timezone to fail")
	}
}

func TestValidateValueHostname(t *testing.T) {
	if err := setting.ValidateValue(setting.KeyHostname, "vyntrio"); err != nil {
		t.Fatalf("ValidateValue hostname: %v", err)
	}
	if err := setting.ValidateValue(setting.KeyHostname, "bad host"); err == nil {
		t.Fatal("expected hostname with space to fail")
	}
}

func TestSettingValidate(t *testing.T) {
	s := setting.Setting{
		Namespace: setting.NamespaceSystem,
		Key:       setting.KeyHostname,
		Value:     "vyntrio",
		ValueType: setting.ValueTypeString,
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	s.ValueType = setting.ValueTypeInt
	if err := s.Validate(); !errors.Is(err, setting.ErrInvalidValueType) {
		t.Fatalf("expected ErrInvalidValueType, got %v", err)
	}
}

func TestValidateValueHostnameMaxLength(t *testing.T) {
	long := strings.Repeat("a", 254)
	if err := setting.ValidateValue(setting.KeyHostname, long); err == nil {
		t.Fatal("expected overlong hostname to fail")
	}
}
