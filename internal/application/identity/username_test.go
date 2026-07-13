package identity_test

import (
	"errors"
	"testing"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
)

func TestValidateBootstrapUsername(t *testing.T) {
	if err := appidentity.ValidateBootstrapUsername("owner"); err != nil {
		t.Fatalf("valid username error: %v", err)
	}
	if err := appidentity.ValidateBootstrapUsername(""); !errors.Is(err, appidentity.ErrInvalidBootstrapUsername) {
		t.Fatalf("empty username error = %v", err)
	}
	if err := appidentity.ValidateBootstrapUsername(" owner"); !errors.Is(err, appidentity.ErrInvalidBootstrapUsername) {
		t.Fatalf("leading space error = %v", err)
	}
}
