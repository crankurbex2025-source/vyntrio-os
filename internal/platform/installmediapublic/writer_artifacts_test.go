package installmediapublic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWriterArtifactsEmptyDir(t *testing.T) {
	artifacts := loadWriterArtifacts(t.TempDir())
	if len(artifacts) != len(WriterArtifactNames) {
		t.Fatalf("len = %d want %d", len(artifacts), len(WriterArtifactNames))
	}
	for _, artifact := range artifacts {
		if artifact.DownloadAvailable {
			t.Fatalf("artifact %s should not be available", artifact.Name)
		}
		if artifact.DownloadPath != "" {
			t.Fatalf("artifact %s should have no download path", artifact.Name)
		}
	}
}

func TestLoadWriterArtifactsStaged(t *testing.T) {
	dir := t.TempDir()
	writerDir := filepath.Join(dir, WriterStagingSubdir)
	if err := os.MkdirAll(writerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "vyntrio-media-creator-windows-amd64-setup.exe"
	if err := os.WriteFile(filepath.Join(writerDir, name), []byte("binary-bytes"), 0o755); err != nil {
		t.Fatal(err)
	}
	digest := strings.Repeat("ab", 32)
	sidecar := digest + "  " + name + "\n"
	if err := os.WriteFile(filepath.Join(writerDir, name+".sha256"), []byte(sidecar), 0o644); err != nil {
		t.Fatal(err)
	}

	artifacts := loadWriterArtifacts(dir)
	var found *WriterArtifact
	for i := range artifacts {
		if artifacts[i].Name == name {
			found = &artifacts[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("windows artifact not reported")
	}
	if !found.DownloadAvailable {
		t.Fatal("windows artifact should be available")
	}
	if found.DownloadPath != "/release/writer/"+name {
		t.Fatalf("download path = %q", found.DownloadPath)
	}
	if found.SHA256 != digest {
		t.Fatalf("sha256 = %q want %q", found.SHA256, digest)
	}
	if found.SizeBytes == nil || *found.SizeBytes != uint64(len("binary-bytes")) {
		t.Fatalf("size = %v", found.SizeBytes)
	}
}
