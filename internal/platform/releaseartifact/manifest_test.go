package releaseartifact_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
)

func TestDecodeManifestValid(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"bin/vyntrio-api": "payload",
	})
	manifestPath := filepath.Join(dir, "release-manifest.json")
	data := validManifestBytes(t, dir, "bin/vyntrio-api", "payload")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got, err := releaseartifact.DecodeManifest(data)
	if err != nil {
		t.Fatalf("DecodeManifest() error: %v", err)
	}
	if got.FormatVersion != releaseartifact.FormatVersion {
		t.Fatalf("format_version = %q", got.FormatVersion)
	}
	if len(got.Artifacts) != 1 {
		t.Fatalf("len(artifacts) = %d", len(got.Artifacts))
	}
}

func TestDecodeManifestRejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{
			name: "unknown_top_level_field",
			json: `{"format_version":"vyntrio-release-manifest-v1","created_at":"2026-07-15T12:00:00Z","release":{"version":"1.0.0"},"artifacts":[{"name":"a","type":"binary","relative_path":"a.bin","size_bytes":1,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","use":"local_verification"}],"extra":true}`,
		},
		{
			name: "unsupported_format_version",
			json: `{"format_version":"vyntrio-release-manifest-v0","created_at":"2026-07-15T12:00:00Z","release":{"version":"1.0.0"},"artifacts":[{"name":"a","type":"binary","relative_path":"a.bin","size_bytes":1,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","use":"local_verification"}]}`,
		},
		{
			name: "unknown_artifact_type",
			json: `{"format_version":"vyntrio-release-manifest-v1","created_at":"2026-07-15T12:00:00Z","release":{"version":"1.0.0"},"artifacts":[{"name":"a","type":"docker_image","relative_path":"a.bin","size_bytes":1,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","use":"local_verification"}]}`,
		},
		{
			name: "unknown_artifact_use",
			json: `{"format_version":"vyntrio-release-manifest-v1","created_at":"2026-07-15T12:00:00Z","release":{"version":"1.0.0"},"artifacts":[{"name":"a","type":"binary","relative_path":"a.bin","size_bytes":1,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","use":"usb_flash"}]}`,
		},
		{
			name: "path_traversal",
			json: `{"format_version":"vyntrio-release-manifest-v1","created_at":"2026-07-15T12:00:00Z","release":{"version":"1.0.0"},"artifacts":[{"name":"a","type":"binary","relative_path":"../escape.bin","size_bytes":1,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","use":"local_verification"}]}`,
		},
		{
			name: "invalid_sha256",
			json: `{"format_version":"vyntrio-release-manifest-v1","created_at":"2026-07-15T12:00:00Z","release":{"version":"1.0.0"},"artifacts":[{"name":"a","type":"binary","relative_path":"a.bin","size_bytes":1,"sha256":"not-a-hash","use":"local_verification"}]}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := releaseartifact.DecodeManifest([]byte(tc.json))
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestVerifyValidManifestAndArtifact(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"bin/vyntrio-api": "payload-bytes",
	})
	manifestPath := filepath.Join(dir, "release-manifest.json")
	data := validManifestBytes(t, dir, "bin/vyntrio-api", "payload-bytes")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, dir)
	if err != nil {
		t.Fatalf("VerifyManifestFile() error: %v", err)
	}
	if result.Integrity != releaseartifact.IntegrityOK {
		t.Fatalf("integrity = %q", result.Integrity)
	}
	if result.Authenticity != releaseartifact.AuthenticityNotSigned {
		t.Fatalf("authenticity = %q", result.Authenticity)
	}
}

func TestVerifyMissingArtifactFails(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "release-manifest.json")
	data := validManifestBytes(t, dir, "bin/missing", "payload")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifySizeMismatchFails(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"bin/vyntrio-api": "short",
	})
	manifestPath := filepath.Join(dir, "release-manifest.json")
	data := validManifestBytes(t, dir, "bin/vyntrio-api", "much-longer-than-short")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if len(result.Failures) != 1 || result.Failures[0].Reason != releaseartifact.ReasonSizeMismatch {
		t.Fatalf("failures = %+v", result.Failures)
	}
}

func TestVerifySHA256MismatchFails(t *testing.T) {
	content := "payload-bytes"
	dir := writeFixture(t, map[string]string{
		"bin/vyntrio-api": content,
	})
	manifestPath := filepath.Join(dir, "release-manifest.json")
	manifest := map[string]any{
		"format_version": releaseartifact.FormatVersion,
		"created_at":     "2026-07-15T12:00:00Z",
		"release":        map[string]string{"version": "0.2.0-dev"},
		"artifacts": []map[string]any{
			{
				"name":          "vyntrio-api",
				"type":          "binary",
				"relative_path": "bin/vyntrio-api",
				"size_bytes":    len(content),
				"sha256":        strings.Repeat("a", 64),
				"use":           "local_verification",
			},
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if len(result.Failures) != 1 || result.Failures[0].Reason != releaseartifact.ReasonSHA256Mismatch {
		t.Fatalf("failures = %+v", result.Failures)
	}
}

func TestVerifySignaturePresentReportsUnsupportedAuthenticity(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"bin/vyntrio-api": "payload-bytes",
	})
	manifest := map[string]any{
		"format_version": releaseartifact.FormatVersion,
		"created_at":     "2026-07-15T12:00:00Z",
		"release":        map[string]string{"version": "0.2.0-dev", "channel": "development"},
		"artifacts": []map[string]any{
			artifactEntry(t, "bin/vyntrio-api", "payload-bytes"),
		},
		"signature": map[string]string{
			"algorithm": "ed25519",
			"key_id":    "vyntrio-release-v1",
			"value":     "c2lnbmF0dXJl",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	manifestPath := filepath.Join(dir, "release-manifest.json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, dir)
	if err != nil {
		t.Fatalf("VerifyManifestFile() error: %v", err)
	}
	if result.Authenticity != releaseartifact.AuthenticityUnsupported {
		t.Fatalf("authenticity = %q", result.Authenticity)
	}
}

func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}
	}
	return dir
}

func validManifestBytes(t *testing.T, dir, relPath, content string) []byte {
	t.Helper()
	manifest := map[string]any{
		"format_version": releaseartifact.FormatVersion,
		"created_at":     "2026-07-15T12:00:00Z",
		"release":        map[string]string{"version": "0.2.0-dev", "channel": "development"},
		"artifacts": []map[string]any{
			artifactEntry(t, relPath, content),
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	_ = dir
	return data
}

func artifactEntry(t *testing.T, relPath, content string) map[string]any {
	t.Helper()
	sum := sha256.Sum256([]byte(content))
	return map[string]any{
		"name":          strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath)),
		"type":          "binary",
		"relative_path": filepath.ToSlash(relPath),
		"size_bytes":    len(content),
		"sha256":        hex.EncodeToString(sum[:]),
		"use":           "local_verification",
	}
}
