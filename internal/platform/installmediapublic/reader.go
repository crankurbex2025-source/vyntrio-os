package installmediapublic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
)

// Reader loads public install-media metadata from a release staging directory.
type Reader struct {
	StagingDir string
	Version    string
	Commit     string
}

// Read returns metadata for the public install-media endpoint.
func (r Reader) Read() (Metadata, error) {
	stagingDir := strings.TrimSpace(r.StagingDir)
	if stagingDir == "" {
		return notBuiltMetadata(r.Version), nil
	}

	publicPath := filepath.Join(stagingDir, PublicMetadataName)
	data, err := os.ReadFile(publicPath)
	if err != nil {
		if os.IsNotExist(err) {
			return notBuiltMetadata(r.Version), nil
		}
		return Metadata{}, fmt.Errorf("read public metadata: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return Metadata{}, fmt.Errorf("decode public metadata: %w", err)
	}
	if err := validateLoadedMetadata(metadata); err != nil {
		return Metadata{}, err
	}

	artifactPath := filepath.Join(stagingDir, metadata.PrimaryArtifact.Name)
	info, err := os.Stat(artifactPath)
	if err != nil {
		metadata.PublicationStatus = PublicationUnavailable
		metadata.PrimaryArtifact.DownloadAvailable = false
		metadata.PrimaryArtifact.DownloadPath = ""
		metadata.PrimaryArtifact.ManifestPath = ""
		metadata.PrimaryArtifact.SizeBytes = nil
		metadata.PrimaryArtifact.SHA256 = ""
		return metadata, nil
	}

	size := uint64(info.Size())
	metadata.PrimaryArtifact.SizeBytes = &size
	metadata.PublicationStatus = PublicationLocalStaging
	metadata.PrimaryArtifact.DownloadAvailable = true
	if metadata.PrimaryArtifact.DownloadPath == "" {
		metadata.PrimaryArtifact.DownloadPath = "/release/" + metadata.PrimaryArtifact.Name
	}
	if metadata.PrimaryArtifact.ManifestPath == "" {
		metadata.PrimaryArtifact.ManifestPath = "/release/" + ManifestName
	}

	manifestPath := filepath.Join(stagingDir, ManifestName)
	if _, err := os.Stat(manifestPath); err == nil {
		if _, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, stagingDir); err != nil {
			metadata.PublicationStatus = PublicationUnavailable
			metadata.PrimaryArtifact.DownloadAvailable = false
		}
	}

	if metadata.Writer == nil {
		metadata.Writer = defaultWriterInfo()
	} else {
		// Prefer native Tauri metadata even when staged JSON is older.
		metadata.Writer.Name = "vyntrio-media-creator"
		metadata.Writer.BinaryName = "vyntrio-media-creator"
		metadata.Writer.Kind = "native_desktop_tauri"
		metadata.Writer.GUIAvailable = true
		metadata.Writer.GUIKind = "tauri"
		metadata.Writer.NativeGUI = true
		metadata.Writer.BuildTarget = "make build-media-creator-native"
		metadata.Writer.PackageTarget = "make package-media-creator-native"
		metadata.Writer.Platforms = []string{"linux", "windows"}
	}
	metadata.Writer.Artifacts = loadWriterArtifacts(stagingDir)

	return metadata, nil
}

// loadWriterArtifacts inspects the staged writer directory and reports each
// known platform binary with its real size and SHA-256, or as unavailable.
func loadWriterArtifacts(stagingDir string) []WriterArtifact {
	writerDir := filepath.Join(stagingDir, WriterStagingSubdir)
	artifacts := make([]WriterArtifact, 0, len(WriterArtifactNames))
	for _, template := range WriterArtifactNames {
		artifact := template
		path := filepath.Join(writerDir, artifact.Name)
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			artifacts = append(artifacts, artifact)
			continue
		}
		size := uint64(info.Size())
		artifact.SizeBytes = &size
		artifact.SHA256 = readSidecarSHA256(path + ".sha256")
		artifact.DownloadAvailable = true
		artifact.DownloadPath = "/release/writer/" + artifact.Name
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

// readSidecarSHA256 parses "digest  filename" sidecar files written at package time.
func readSidecarSHA256(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return ""
	}
	digest := strings.ToLower(strings.TrimSpace(fields[0]))
	if len(digest) != 64 {
		return ""
	}
	return digest
}

func notBuiltMetadata(version string) Metadata {
	if strings.TrimSpace(version) == "" {
		version = "0.2.0-dev"
	}
	return Metadata{
		PublicationStatus: PublicationNotBuilt,
		Release: ReleaseLine{
			Version: version,
			Channel: "development",
		},
		PrimaryArtifact: PrimaryArtifact{
			Name:              ArtifactName,
			Format:            "raw_gpt_hybrid_disk",
			FirmwareBootMode:  "bios+uefi",
			BiosSupport:       true,
			UefiSupport:       true,
			DualMode:          true,
			DownloadAvailable: false,
		},
		BuildTarget: "make install-media",
		StageTarget: "make release-install-media-stage",
		Limitations: defaultLimitations(),
		SupportStatus: SupportStatusEngineering,
		Writer:      defaultWriterInfo(),
	}
}

func defaultLimitations() []string {
	return []string{
		"Dual-mode raw GPT image required: BIOS/legacy + UEFI (ESP/BOOTX64.EFI)",
		"BIOS-only media is incomplete and must not be treated as the product baseline",
		"Secure Boot is not signed for this engineering image",
		"Runtime boot and dashboard reachability are not proven until a VM/hardware harness reports success",
		"No production CDN — download appears only after local release staging on the serving host",
		"Target-disk installer service is not started — live image boots to RAM dashboard, not a full disk install",
		"Ed25519 release signatures are not verified in v1",
	}
}

func validateLoadedMetadata(metadata Metadata) error {
	if strings.TrimSpace(metadata.PrimaryArtifact.Name) == "" {
		return fmt.Errorf("primary artifact name is required")
	}
	if metadata.PrimaryArtifact.Name != ArtifactName && metadata.PrimaryArtifact.Name != ArtifactNameLegacyBIOS {
		return fmt.Errorf("unexpected primary artifact name %q", metadata.PrimaryArtifact.Name)
	}
	if strings.TrimSpace(metadata.PrimaryArtifact.Format) == "" {
		return fmt.Errorf("primary artifact format is required")
	}
	if metadata.GeneratedAt != "" {
		if _, err := time.Parse(time.RFC3339Nano, metadata.GeneratedAt); err != nil {
			if _, err2 := time.Parse(time.RFC3339, metadata.GeneratedAt); err2 != nil {
				return fmt.Errorf("generated_at is not RFC3339")
			}
		}
	}
	return nil
}
