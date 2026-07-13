package settings_test

import (
	"testing"

	appsettings "github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
)

func TestPublicViewResponse(t *testing.T) {
	snapshot := appsettings.NewSnapshot(appsettings.SystemSettings{
		Timezone: "UTC",
		Hostname: "vyntrio",
	})
	view := appsettings.NewPublicView(snapshot, "0.2.0-dev", "development")

	got := view.Response()
	if got.Instance.Name != "vyntrio" {
		t.Fatalf("instance.name = %q, want vyntrio", got.Instance.Name)
	}
	if got.Instance.Version != "0.2.0-dev" {
		t.Fatalf("instance.version = %q, want 0.2.0-dev", got.Instance.Version)
	}
	if got.API.Environment != "development" {
		t.Fatalf("api.environment = %q, want development", got.API.Environment)
	}
}
