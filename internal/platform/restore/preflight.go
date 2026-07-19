package restore

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

// PreflightArchive performs manifest-first validation with zero live-state mutation.
func PreflightArchive(path string) (backup.Manifest, error) {
	manifest, err := readEmbeddedManifest(path)
	if err != nil {
		return backup.Manifest{}, err
	}
	if err := backup.ValidateArchive(path, manifest); err != nil {
		return backup.Manifest{}, fmt.Errorf("%w: %v", ErrArtifactInvalid, err)
	}
	if err := CheckCompatibility(manifest); err != nil {
		return backup.Manifest{}, err
	}
	return manifest, nil
}

// ExtractRestorableMembers reads validated archive payloads for placement.
// PreflightArchive must succeed before calling this function.
func ExtractRestorableMembers(path string, manifest backup.Manifest) (map[string][]byte, error) {
	if err := CheckCompatibility(manifest); err != nil {
		return nil, err
	}
	expected := make(map[string]backup.ManifestMember, len(manifest.Members))
	for _, member := range manifest.Members {
		expected[member.Name] = member
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: open artifact", ErrArtifactInvalid)
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	payloads := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: read tar", ErrArtifactInvalid)
		}
		if hdr.Name == backup.ManifestFileName {
			continue
		}
		want, ok := expected[hdr.Name]
		if !ok {
			return nil, fmt.Errorf("%w: unexpected member %q", ErrArtifactInvalid, hdr.Name)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("%w: read member", ErrArtifactInvalid)
		}
		if int64(len(data)) != want.Size {
			return nil, fmt.Errorf("%w: member size mismatch", ErrArtifactInvalid)
		}
		payloads[hdr.Name] = data
	}
	if _, ok := payloads[backup.StateDBMember]; !ok {
		return nil, fmt.Errorf("%w: required state member missing", ErrArtifactInvalid)
	}
	return payloads, nil
}

func readEmbeddedManifest(path string) (backup.Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return backup.Manifest{}, fmt.Errorf("%w: open artifact", ErrArtifactInvalid)
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return backup.Manifest{}, fmt.Errorf("%w: manifest missing", ErrArtifactInvalid)
		}
		if err != nil {
			return backup.Manifest{}, fmt.Errorf("%w: read tar", ErrArtifactInvalid)
		}
		if hdr.Name != backup.ManifestFileName {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return backup.Manifest{}, fmt.Errorf("%w: read manifest", ErrArtifactInvalid)
		}
		manifest, err := backup.DecodeManifest(data)
		if err != nil {
			return backup.Manifest{}, fmt.Errorf("%w: %v", ErrArtifactInvalid, err)
		}
		return manifest, nil
	}
}
