package overview_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
)

func TestAssembleSoftwareOKWithCommit(t *testing.T) {
	got := overview.AssembleSoftware("0.2.0-dev", "abc123", "development")
	if got.Status != overview.SoftwareStatusOK {
		t.Fatalf("status = %q, want ok", got.Status)
	}
	if got.Version != "0.2.0-dev" || got.Commit != "abc123" || got.Channel != overview.ReleaseChannelDevelopment {
		t.Fatalf("software = %+v", got)
	}
}

func TestAssembleSoftwareOKWithoutCommit(t *testing.T) {
	got := overview.AssembleSoftware("0.2.0-dev", "", "production")
	if got.Status != overview.SoftwareStatusOK {
		t.Fatalf("status = %q, want ok", got.Status)
	}
	if got.Commit != "" {
		t.Fatalf("commit = %q, want omitted", got.Commit)
	}
	if got.Channel != overview.ReleaseChannelProduction {
		t.Fatalf("channel = %q", got.Channel)
	}
}

func TestAssembleSoftwareUnavailableWithoutVersion(t *testing.T) {
	got := overview.AssembleSoftware("  ", "abc123", "development")
	if got.Status != overview.SoftwareStatusUnavailable {
		t.Fatalf("status = %q, want unavailable", got.Status)
	}
	if got.Version != "" || got.Commit != "" || got.Channel != "" {
		t.Fatalf("software = %+v, want status only", got)
	}
}

func TestMapReleaseChannelUnknown(t *testing.T) {
	if got := overview.MapReleaseChannel("staging"); got != overview.ReleaseChannelUnknown {
		t.Fatalf("channel = %q, want unknown", got)
	}
}
