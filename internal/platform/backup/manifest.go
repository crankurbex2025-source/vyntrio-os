package backup

import (
	"encoding/json"
	"fmt"
	"time"
)

// ManifestMember describes one non-manifest archive member.
type ManifestMember struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// Manifest is the v1 backup manifest archived as manifest.json.
// Manifest integrity is not cryptographically authenticated in v1.
type Manifest struct {
	FormatVersion    string           `json:"format_version"`
	CreatedAt        string           `json:"created_at"`
	APIVersion       string           `json:"api_version,omitempty"`
	APICommit        string           `json:"api_commit,omitempty"`
	MigrationVersion string           `json:"migration_version,omitempty"`
	MigrationDirty   *bool            `json:"migration_dirty,omitempty"`
	Members          []ManifestMember `json:"members"`
}

func NewManifest(createdAt time.Time, metadata ReleaseMetadata, members []ManifestMember) Manifest {
	m := Manifest{
		FormatVersion: FormatVersion,
		CreatedAt:     createdAt.UTC().Format(time.RFC3339Nano),
		Members:       members,
	}
	if metadata.APIVersion != "" {
		m.APIVersion = metadata.APIVersion
	}
	if metadata.APICommit != "" {
		m.APICommit = metadata.APICommit
	}
	if metadata.MigrationVersion != "" {
		m.MigrationVersion = metadata.MigrationVersion
		if metadata.MigrationDirtyKnown {
			dirty := metadata.MigrationDirty
			m.MigrationDirty = &dirty
		}
	}
	return m
}

func (m Manifest) Encode() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	return data, nil
}

func DecodeManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return m, nil
}

// ReleaseMetadata holds optional manifest metadata fields.
type ReleaseMetadata struct {
	APIVersion          string
	APICommit           string
	MigrationVersion    string
	MigrationDirty      bool
	MigrationDirtyKnown bool
}
