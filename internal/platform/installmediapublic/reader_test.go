package installmediapublic_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installmediapublic"
)

func TestReaderNotBuiltWhenStagingMissing(t *testing.T) {
	got, err := installmediapublic.Reader{Version: "0.2.0-dev"}.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got.PublicationStatus != installmediapublic.PublicationNotBuilt {
		t.Fatalf("status = %q", got.PublicationStatus)
	}
	if got.PrimaryArtifact.DownloadAvailable {
		t.Fatal("expected download unavailable")
	}
	if got.PrimaryArtifact.Name != installmediapublic.ArtifactName {
		t.Fatalf("name = %q", got.PrimaryArtifact.Name)
	}
	if !got.PrimaryArtifact.UefiSupport || !got.PrimaryArtifact.DualMode || !got.PrimaryArtifact.BiosSupport {
		t.Fatalf("expected dual-mode bios+uefi defaults, got bios=%v uefi=%v dual=%v",
			got.PrimaryArtifact.BiosSupport, got.PrimaryArtifact.UefiSupport, got.PrimaryArtifact.DualMode)
	}
}

func TestReaderLocalStagingFromGeneratedMetadata(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, installmediapublic.ArtifactName)
	if err := os.WriteFile(artifactPath, []byte("fake-install-image"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	public := installmediapublic.Metadata{
		PublicationStatus: installmediapublic.PublicationLocalStaging,
		GeneratedAt:       "2026-07-16T12:00:00Z",
		Release: installmediapublic.ReleaseLine{
			Version: "0.2.0-dev",
			Channel: "development",
		},
		PrimaryArtifact: installmediapublic.PrimaryArtifact{
			Name:             installmediapublic.ArtifactName,
			Format:           "raw_gpt_hybrid_disk",
			FirmwareBootMode: "bios+uefi",
			DownloadPath:     "/release/" + installmediapublic.ArtifactName,
			ManifestPath:     "/release/" + installmediapublic.ManifestName,
		},
		BuildTarget: "make install-media",
		StageTarget: "make release-install-media-stage",
	}
	data, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, installmediapublic.PublicMetadataName), data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got, err := installmediapublic.Reader{StagingDir: dir, Version: "0.2.0-dev"}.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got.PublicationStatus != installmediapublic.PublicationLocalStaging {
		t.Fatalf("status = %q", got.PublicationStatus)
	}
	if !got.PrimaryArtifact.DownloadAvailable {
		t.Fatal("expected download available")
	}
	if got.PrimaryArtifact.SizeBytes == nil || *got.PrimaryArtifact.SizeBytes != 18 {
		t.Fatalf("size_bytes = %v", got.PrimaryArtifact.SizeBytes)
	}
}
