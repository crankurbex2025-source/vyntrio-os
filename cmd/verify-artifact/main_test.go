package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunUsage(t *testing.T) {
	if run(nil) != 2 {
		t.Fatal("expected usage exit code 2 for empty args")
	}
	if run([]string{"help"}) != 0 {
		t.Fatal("expected help exit code 0")
	}
	if run([]string{"--base-dir", "/tmp"}) != 2 {
		t.Fatal("expected usage exit code 2 when manifest missing")
	}
}

func TestRunMissingManifestFails(t *testing.T) {
	if run([]string{"/does/not/exist/manifest.json"}) != 1 {
		t.Fatal("expected failure for missing manifest")
	}
}

func TestRunValidFixtureSucceeds(t *testing.T) {
	dir := t.TempDir()
	rel := "bin/vyntrio-api"
	content := []byte("fixture-payload")
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	manifestPath := filepath.Join(dir, "release-manifest.json")
	manifest := map[string]any{
		"format_version": "vyntrio-release-manifest-v1",
		"created_at":     "2026-07-15T12:00:00Z",
		"release":        map[string]string{"version": "0.2.0-dev"},
		"artifacts": []map[string]any{
			{
				"name":          "vyntrio-api",
				"type":          "binary",
				"relative_path": rel,
				"size_bytes":    len(content),
				"sha256":        sha256Hex(content),
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

	if run([]string{manifestPath}) != 0 {
		t.Fatal("expected success")
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestRunMalformedManifestFails(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/manifest.json"
	if err := os.WriteFile(path, []byte(`{"format_version":"bad"}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	if run([]string{path}) != 1 {
		t.Fatal("expected failure for malformed manifest")
	}
}
