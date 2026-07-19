package writemedia

import (
	"fmt"
	"os"
	"strings"
)

// VerifyImageFile recomputes SHA-256 for an image file and compares to expected when set.
func VerifyImageFile(imagePath, expectedSHA string) error {
	img, err := LoadImage(imagePath)
	if err != nil {
		return err
	}
	expected := strings.ToLower(strings.TrimSpace(expectedSHA))
	if expected == "" {
		expected = img.ExpectedSHA256
	}
	if expected == "" {
		fmt.Fprintf(os.Stderr, "verify-image: SHA-256=%s (no expected digest in manifest)\n", img.SHA256)
		return nil
	}
	if !strings.EqualFold(img.SHA256, expected) {
		return fmt.Errorf("SHA-256 mismatch: file=%s expected=%s", img.SHA256, expected)
	}
	fmt.Fprintf(os.Stderr, "verify-image: SHA-256 ok (%s)\n", img.SHA256)
	return nil
}

// VerifyDevice reads back a device and compares SHA-256 to the source image when sizes allow.
func VerifyDevice(imagePath, devicePath string) error {
	img, err := LoadImage(imagePath)
	if err != nil {
		return err
	}
	devicePath = strings.TrimSpace(devicePath)
	if devicePath == "" {
		return fmt.Errorf("device path is required")
	}
	deviceSHA, err := hashDevice(devicePath, img.SizeBytes)
	if err != nil {
		return err
	}
	if !strings.EqualFold(deviceSHA, img.SHA256) {
		return fmt.Errorf("device SHA-256 mismatch: device=%s image=%s", deviceSHA, img.SHA256)
	}
	fmt.Fprintf(os.Stderr, "verify-device: SHA-256 ok (%s)\n", deviceSHA)
	return nil
}
