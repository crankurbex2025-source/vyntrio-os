package writemedia

import "fmt"

// Device is a removable or external block device candidate for install media.
type Device struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	SizeBytes   uint64   `json:"size_bytes"`
	Removable   bool     `json:"removable"`
	BusType     string   `json:"bus_type"`
	Mounted     bool     `json:"mounted"`
	MountPoints []string `json:"mount_points,omitempty"`
}

// Image describes the BIOS raw install image to write.
type Image struct {
	Path           string
	Name           string
	SizeBytes      uint64
	SHA256         string
	ExpectedSHA256 string
	ManifestPath   string
}

// WriteResult summarizes a completed write operation.
type WriteResult struct {
	DevicePath string
	ImagePath  string
	BytesWritten uint64
	Verified   bool
}

// ErrDestructiveWrite is returned when confirmation is required.
var ErrDestructiveWrite = fmt.Errorf("destructive write requires explicit confirmation")
