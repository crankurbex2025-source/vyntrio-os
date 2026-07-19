package writemedia

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installmediapublic"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
)

const manifestName = installmediapublic.ManifestName

// LoadImage loads image metadata and verifies SHA-256 when a manifest is present.
func LoadImage(imagePath string) (Image, error) {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return Image{}, fmt.Errorf("image path is required")
	}
	info, err := os.Stat(imagePath)
	if err != nil {
		return Image{}, fmt.Errorf("stat image: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Image{}, fmt.Errorf("image is not a regular file: %s", imagePath)
	}

	img := Image{
		Path:      imagePath,
		Name:      filepath.Base(imagePath),
		SizeBytes: uint64(info.Size()),
	}

	sha, err := hashFile(imagePath)
	if err != nil {
		return Image{}, err
	}
	img.SHA256 = sha

	manifestPath := filepath.Join(filepath.Dir(imagePath), manifestName)
	if _, err := os.Stat(manifestPath); err != nil {
		return img, nil
	}
	img.ManifestPath = manifestPath

	result, err := releaseartifact.NewVerifier().VerifyManifestFile(manifestPath, filepath.Dir(imagePath))
	if err != nil {
		return Image{}, fmt.Errorf("verify manifest: %w", err)
	}
	_ = result
	expected, err := expectedSHAFromManifest(manifestPath, img.Name)
	if err != nil {
		return Image{}, err
	}
	img.ExpectedSHA256 = expected
	if img.ExpectedSHA256 != "" && !strings.EqualFold(img.ExpectedSHA256, img.SHA256) {
		return Image{}, fmt.Errorf("SHA-256 mismatch: file=%s manifest=%s", img.SHA256, img.ExpectedSHA256)
	}
	return img, nil
}

func expectedSHAFromManifest(manifestPath, imageName string) (string, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", err
	}
	var payload struct {
		Artifacts []struct {
			RelativePath string `json:"relative_path"`
			SHA256       string `json:"sha256"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", fmt.Errorf("decode manifest: %w", err)
	}
	for _, artifact := range payload.Artifacts {
		if artifact.RelativePath == imageName {
			return strings.ToLower(strings.TrimSpace(artifact.SHA256)), nil
		}
	}
	return "", fmt.Errorf("artifact %q not found in manifest", imageName)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
