package releaseartifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Manifest is the v1 release/install artifact inventory.
// Integrity is established via per-artifact SHA-256; cryptographic authenticity
// is not verified unless a future verifier implements the signature block.
type Manifest struct {
	FormatVersion string              `json:"format_version"`
	CreatedAt     string              `json:"created_at"`
	Release       ReleaseInfo         `json:"release"`
	Artifacts     []Artifact          `json:"artifacts"`
	Signature     *SignatureMetadata  `json:"signature,omitempty"`
}

// ReleaseInfo describes the release line the artifacts belong to.
type ReleaseInfo struct {
	Version string `json:"version"`
	Channel string `json:"channel,omitempty"`
	BuildID string `json:"build_id,omitempty"`
}

// Artifact describes one verifiable file relative to the manifest base directory.
type Artifact struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	RelativePath string `json:"relative_path"`
	SizeBytes    uint64 `json:"size_bytes"`
	SHA256       string `json:"sha256"`
	Use          string `json:"use"`
}

// SignatureMetadata reserves future Ed25519 release signing.
// Presence does not imply verification in v1.
type SignatureMetadata struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id,omitempty"`
	Value     string `json:"value"`
}

// DecodeManifest parses and structurally validates a release manifest.
func DecodeManifest(data []byte) (Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("%w: %v", ErrMalformedManifest, err)
	}
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func validateManifest(manifest Manifest) error {
	if manifest.FormatVersion != FormatVersion {
		return failuref(ReasonUnsupportedFormatVersion, "format_version=%q", manifest.FormatVersion)
	}
	if strings.TrimSpace(manifest.CreatedAt) == "" {
		return failuref(ReasonMalformedManifest, "created_at is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, manifest.CreatedAt); err != nil {
		if _, err2 := time.Parse(time.RFC3339, manifest.CreatedAt); err2 != nil {
			return failuref(ReasonMalformedManifest, "created_at is not RFC3339")
		}
	}
	if strings.TrimSpace(manifest.Release.Version) == "" {
		return failuref(ReasonMalformedManifest, "release.version is required")
	}
	if manifest.Release.Channel != "" {
		if _, ok := allowedReleaseChannels[manifest.Release.Channel]; !ok {
			return failuref(ReasonMalformedManifest, "release.channel=%q", manifest.Release.Channel)
		}
	}
	if len(manifest.Artifacts) == 0 {
		return failuref(ReasonMalformedManifest, "artifacts must not be empty")
	}

	seen := make(map[string]struct{}, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		if strings.TrimSpace(artifact.Name) == "" {
			return failuref(ReasonMalformedManifest, "artifact.name is required")
		}
		if _, ok := seen[artifact.Name]; ok {
			return failuref(ReasonDuplicateArtifactName, "name=%s", artifact.Name)
		}
		seen[artifact.Name] = struct{}{}

		if _, ok := allowedArtifactTypes[artifact.Type]; !ok {
			return failuref(ReasonUnknownArtifactType, "name=%s type=%q", artifact.Name, artifact.Type)
		}
		if _, ok := allowedArtifactUses[artifact.Use]; !ok {
			return failuref(ReasonUnknownArtifactUse, "name=%s use=%q", artifact.Name, artifact.Use)
		}
		if err := validateRelativePath(artifact.RelativePath); err != nil {
			return failuref(ReasonInvalidRelativePath, "name=%s %v", artifact.Name, err)
		}
		if artifact.SizeBytes == 0 {
			return failuref(ReasonMalformedManifest, "name=%s size_bytes must be > 0", artifact.Name)
		}
		if !sha256Pattern.MatchString(artifact.SHA256) {
			return failuref(ReasonInvalidSHA256, "name=%s", artifact.Name)
		}
	}
	if manifest.Signature != nil {
		if strings.TrimSpace(manifest.Signature.Algorithm) == "" || strings.TrimSpace(manifest.Signature.Value) == "" {
			return failuref(ReasonMalformedManifest, "signature requires algorithm and value when present")
		}
	}
	return nil
}

func validateRelativePath(relativePath string) error {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return fmt.Errorf("relative_path is required")
	}
	if filepath.IsAbs(relativePath) {
		return fmt.Errorf("relative_path must not be absolute")
	}
	clean := filepath.Clean(relativePath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("relative_path must stay within base directory")
	}
	return nil
}

func failuref(reason, format string, args ...any) error {
	return fmt.Errorf("%w: reason=%s detail=%s", ErrMalformedManifest, reason, fmt.Sprintf(format, args...))
}
