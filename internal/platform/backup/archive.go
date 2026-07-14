package backup

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type memberPayload struct {
	Name string
	Data []byte
}

// BuildArchive writes a validated temporary tar artifact and returns its path.
func BuildArchive(tempPath string, createdAt time.Time, metadata ReleaseMetadata, sources []SourceMember) (Manifest, error) {
	payloads, manifestMembers, err := readSourcePayloads(sources)
	if err != nil {
		return Manifest{}, err
	}

	manifest := NewManifest(createdAt, metadata, manifestMembers)
	manifestData, err := manifest.Encode()
	if err != nil {
		return Manifest{}, err
	}
	payloads = append(payloads, memberPayload{Name: ManifestFileName, Data: manifestData})

	if err := writeTar(tempPath, createdAt, payloads); err != nil {
		return Manifest{}, err
	}
	if err := ValidateArchive(tempPath, manifest); err != nil {
		_ = os.Remove(tempPath)
		return Manifest{}, err
	}
	return manifest, nil
}

func readSourcePayloads(sources []SourceMember) ([]memberPayload, []ManifestMember, error) {
	payloads := make([]memberPayload, 0, len(sources))
	members := make([]ManifestMember, 0, len(sources))
	for _, source := range sources {
		if err := RevalidateSourceMember(source); err != nil {
			return nil, nil, err
		}
		data, err := os.ReadFile(source.SourcePath)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: read source", ErrSourceInvalid)
		}
		sum := sha256.Sum256(data)
		payloads = append(payloads, memberPayload{Name: source.ArchiveName, Data: data})
		members = append(members, ManifestMember{
			Name:   source.ArchiveName,
			Size:   int64(len(data)),
			SHA256: hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(members, func(i, j int) bool { return members[i].Name < members[j].Name })
	sort.Slice(payloads, func(i, j int) bool { return payloads[i].Name < payloads[j].Name })
	return payloads, members, nil
}

func writeTar(path string, createdAt time.Time, payloads []memberPayload) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, artifactMode)
	if err != nil {
		return fmt.Errorf("%w: create temporary artifact", ErrArtifactFailed)
	}
	defer func() { _ = f.Close() }()

	tw := tar.NewWriter(f)
	for _, payload := range payloads {
		if err := ValidateArchiveMemberName(payload.Name); err != nil {
			return err
		}
		hdr := &tar.Header{
			Name:     payload.Name,
			Mode:     tarMemberMode,
			Size:     int64(len(payload.Data)),
			ModTime:  createdAt.UTC(),
			Typeflag: tar.TypeReg,
			Uid:      0,
			Gid:      0,
			Uname:    "root",
			Gname:    "root",
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("%w: write tar header", ErrArtifactFailed)
		}
		if _, err := tw.Write(payload.Data); err != nil {
			return fmt.Errorf("%w: write tar body", ErrArtifactFailed)
		}
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("%w: finalize tar", ErrArtifactFailed)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("%w: sync temporary artifact", ErrArtifactFailed)
	}
	return nil
}

// ValidateArchive verifies member names/types and manifest digests.
func ValidateArchive(path string, manifest Manifest) error {
	if manifest.FormatVersion != FormatVersion {
		return fmt.Errorf("%w: unsupported manifest format", ErrArtifactFailed)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%w: open artifact", ErrArtifactFailed)
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	seen := make(map[string]struct{})
	var manifestBytes []byte
	expected := manifestMemberMap(manifest)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: read tar", ErrArtifactFailed)
		}
		if err := validateArchiveHeader(hdr); err != nil {
			return err
		}
		if _, dup := seen[hdr.Name]; dup {
			return fmt.Errorf("%w: duplicate archive member", ErrArtifactFailed)
		}
		seen[hdr.Name] = struct{}{}

		data, err := io.ReadAll(tr)
		if err != nil {
			return fmt.Errorf("%w: read member", ErrArtifactFailed)
		}
		if int64(len(data)) != hdr.Size {
			return fmt.Errorf("%w: member size mismatch", ErrArtifactFailed)
		}

		if hdr.Name == ManifestFileName {
			manifestBytes = append([]byte(nil), data...)
			continue
		}
		want, ok := expected[hdr.Name]
		if !ok {
			return fmt.Errorf("%w: unexpected archive member", ErrArtifactFailed)
		}
		if int64(len(data)) != want.Size {
			return fmt.Errorf("%w: manifest size mismatch", ErrArtifactFailed)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want.SHA256 {
			return fmt.Errorf("%w: digest mismatch", ErrArtifactFailed)
		}
	}

	if len(manifestBytes) == 0 {
		return fmt.Errorf("%w: manifest missing", ErrArtifactFailed)
	}
	decoded, err := DecodeManifest(manifestBytes)
	if err != nil {
		return err
	}
	if !manifestsEquivalent(manifest, decoded) {
		return fmt.Errorf("%w: manifest content mismatch", ErrArtifactFailed)
	}
	if err := validateSeenMembers(seen); err != nil {
		return err
	}
	return nil
}

func validateArchiveHeader(hdr *tar.Header) error {
	if err := ValidateArchiveMemberName(hdr.Name); err != nil {
		return err
	}
	switch hdr.Typeflag {
	case tar.TypeReg, tar.TypeRegA:
	default:
		return fmt.Errorf("%w: non-regular tar entry", ErrArtifactFailed)
	}
	if hdr.Linkname != "" {
		return fmt.Errorf("%w: tar link entry rejected", ErrArtifactFailed)
	}
	return nil
}

// ValidateArchiveMemberName enforces fixed relative archive member names.
func ValidateArchiveMemberName(name string) error {
	if name == "" || name != path.Clean(name) {
		return fmt.Errorf("%w: malformed archive member name", ErrArtifactFailed)
	}
	if strings.HasPrefix(name, "/") || strings.Contains(name, `\`) {
		return fmt.Errorf("%w: absolute archive member rejected", ErrArtifactFailed)
	}
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if part == ".." || part == "" {
			return fmt.Errorf("%w: traversal archive member rejected", ErrArtifactFailed)
		}
	}
	if _, ok := AllowedArchiveMembers[name]; !ok {
		return fmt.Errorf("%w: disallowed archive member", ErrArtifactFailed)
	}
	return nil
}

func validateSeenMembers(seen map[string]struct{}) error {
	if _, ok := seen[ManifestFileName]; !ok {
		return fmt.Errorf("%w: manifest missing", ErrArtifactFailed)
	}
	if _, ok := seen[StateDBMember]; !ok {
		return fmt.Errorf("%w: required state member missing", ErrArtifactFailed)
	}
	if _, ok := seen[ConfigMember]; !ok {
		return fmt.Errorf("%w: required config member missing", ErrArtifactFailed)
	}
	for name := range seen {
		if _, ok := AllowedArchiveMembers[name]; !ok {
			return fmt.Errorf("%w: unexpected archive member", ErrArtifactFailed)
		}
	}
	return nil
}

func manifestMemberMap(m Manifest) map[string]ManifestMember {
	out := make(map[string]ManifestMember, len(m.Members))
	for _, member := range m.Members {
		out[member.Name] = member
	}
	return out
}

func manifestsEquivalent(a, b Manifest) bool {
	if a.FormatVersion != b.FormatVersion || a.CreatedAt != b.CreatedAt {
		return false
	}
	if a.APIVersion != b.APIVersion || a.APICommit != b.APICommit {
		return false
	}
	if a.MigrationVersion != b.MigrationVersion {
		return false
	}
	if !migrationDirtyEqual(a.MigrationDirty, b.MigrationDirty) {
		return false
	}
	if len(a.Members) != len(b.Members) {
		return false
	}
	aa := append([]ManifestMember(nil), a.Members...)
	bb := append([]ManifestMember(nil), b.Members...)
	sort.Slice(aa, func(i, j int) bool { return aa[i].Name < aa[j].Name })
	sort.Slice(bb, func(i, j int) bool { return bb[i].Name < bb[j].Name })
	ab, err := json.Marshal(aa)
	if err != nil {
		return false
	}
	bbJSON, err := json.Marshal(bb)
	if err != nil {
		return false
	}
	return bytes.Equal(ab, bbJSON)
}

func migrationDirtyEqual(a, b *bool) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}
