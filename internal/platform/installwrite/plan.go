package installwrite

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpreflight"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
)

func buildCopyPlan(envelopeRoot, manifestPath, artifactBaseDir string) ([]CopyEntry, string, error) {
	required := installpreflight.RequiredPayloadRelativePaths()
	if strings.TrimSpace(envelopeRoot) == "" && strings.TrimSpace(manifestPath) == "" {
		return nil, "", ErrArtifactSourceRequired
	}

	var manifest releaseartifact.Manifest
	manifestLoaded := false
	if strings.TrimSpace(manifestPath) != "" {
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, "", fmt.Errorf("read release manifest: %w", err)
		}
		manifest, err = releaseartifact.DecodeManifest(data)
		if err != nil {
			return nil, "", err
		}
		manifestLoaded = true
	}

	entries := make([]CopyEntry, 0, len(required))
	for _, targetRel := range required {
		entry, err := planEntry(targetRel, envelopeRoot, manifestLoaded, manifest, manifestPath, artifactBaseDir)
		if err != nil {
			return nil, "", err
		}
		entries = append(entries, entry)
	}
	releaseVersion := ""
	if manifestLoaded {
		releaseVersion = manifest.Release.Version
	}
	return entries, releaseVersion, nil
}

func planEntry(targetRel, envelopeRoot string, manifestLoaded bool, manifest releaseartifact.Manifest, manifestPath, artifactBaseDir string) (CopyEntry, error) {
	mode := payloadModes[targetRel]
	if mode == 0 {
		mode = 0o644
	}

	if manifestLoaded {
		artifact, ok := findInstallArtifact(manifest, targetRel)
		if !ok {
			return CopyEntry{}, fmt.Errorf("%w: missing install artifact for %s", ErrInstallFailed, targetRel)
		}
		base := strings.TrimSpace(artifactBaseDir)
		if base == "" {
			base = filepath.Dir(manifestPath)
		}
		source := filepath.Join(base, filepath.FromSlash(artifact.RelativePath))
		return CopyEntry{
			SourcePath:     source,
			TargetRel:      targetRel,
			Mode:           mode,
			ExpectedSHA256: artifact.SHA256,
			ExpectedSize:   artifact.SizeBytes,
		}, nil
	}

	payloadRoot := installpreflight.PayloadRoot(envelopeRoot)
	source := filepath.Join(payloadRoot, filepath.FromSlash(targetRel))
	return CopyEntry{
		SourcePath: source,
		TargetRel:  targetRel,
		Mode:       mode,
	}, nil
}

func findInstallArtifact(manifest releaseartifact.Manifest, targetRel string) (releaseartifact.Artifact, bool) {
	for _, artifact := range manifest.Artifacts {
		if artifact.Use != "install_media" && artifact.Use != "local_verification" {
			continue
		}
		normalized := normalizeArtifactPath(artifact.RelativePath)
		if normalized == targetRel {
			return artifact, true
		}
	}
	return releaseartifact.Artifact{}, false
}

func normalizeArtifactPath(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "payload/")
	return path
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
