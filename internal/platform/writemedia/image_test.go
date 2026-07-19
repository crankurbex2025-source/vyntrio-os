package writemedia

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestLoadImageWithManifest(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "vyntrio-install-media.img")
	content := []byte("vyntrio-test-image-bytes")
	if err := os.WriteFile(imagePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	sha, err := hashFile(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "format_version": "vyntrio-release-manifest-v1",
  "created_at": "2026-07-16T00:00:00Z",
  "release": { "version": "0.2.0-dev", "channel": "development" },
  "artifacts": [
    {
      "name": "vyntrio-install-media",
      "type": "archive",
      "relative_path": "vyntrio-install-media.img",
      "size_bytes": ` + strconv.Itoa(len(content)) + `,
      "sha256": "` + sha + `",
      "use": "install_media"
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(dir, "release-manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	img, err := LoadImage(imagePath)
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}
	if img.SHA256 != sha {
		t.Fatalf("sha256 = %q want %q", img.SHA256, sha)
	}
	if img.ExpectedSHA256 != sha {
		t.Fatalf("expected sha256 = %q want %q", img.ExpectedSHA256, sha)
	}
}

func TestVerifyImageFileMismatch(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "vyntrio-install-media.img")
	if err := os.WriteFile(imagePath, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyImageFile(imagePath, "deadbeef"); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestWriteImageDryRunRequiresListedDevice(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "vyntrio-install-media.img")
	if err := os.WriteFile(imagePath, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := writemediaWriteDryRun(imagePath, "/dev/no-such-device")
	if err == nil {
		t.Fatal("expected device not found error")
	}
}

func writemediaWriteDryRun(imagePath, devicePath string) (WriteResult, error) {
	return WriteImage(imagePath, devicePath, WriteOptions{DryRun: true})
}
