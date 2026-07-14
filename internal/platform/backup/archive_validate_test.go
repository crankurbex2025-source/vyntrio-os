package backup_test

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

func validManifest(t *testing.T) backup.Manifest {
	t.Helper()
	dbData := []byte("db-bytes")
	cfgData := []byte("cfg-bytes")
	dbSum := sha256.Sum256(dbData)
	cfgSum := sha256.Sum256(cfgData)
	return backup.NewManifest(
		time.Date(2026, 7, 13, 15, 4, 5, 0, time.UTC),
		backup.ReleaseMetadata{},
		[]backup.ManifestMember{
			{Name: backup.ConfigMember, Size: int64(len(cfgData)), SHA256: hex.EncodeToString(cfgSum[:])},
			{Name: backup.StateDBMember, Size: int64(len(dbData)), SHA256: hex.EncodeToString(dbSum[:])},
		},
	)
}

type tarEntry struct {
	name     string
	data     []byte
	typeflag byte
	linkname string
}

func writeMalformedTar(t *testing.T, path string, entries []tarEntry) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	tw := tar.NewWriter(f)
	createdAt := time.Date(2026, 7, 13, 15, 4, 5, 0, time.UTC)
	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		hdr := &tar.Header{
			Name:     entry.name,
			Mode:     0o640,
			Size:     int64(len(entry.data)),
			ModTime:  createdAt,
			Typeflag: typeflag,
			Linkname: entry.linkname,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if len(entry.data) > 0 {
			if _, err := tw.Write(entry.data); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
}

func manifestPayload(t *testing.T, manifest backup.Manifest) []byte {
	t.Helper()
	data, err := manifest.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func memberData(name string, manifest backup.Manifest) []byte {
	for _, member := range manifest.Members {
		if member.Name == name {
			switch name {
			case backup.StateDBMember:
				return []byte("db-bytes")
			case backup.ConfigMember:
				return []byte("cfg-bytes")
			}
		}
	}
	return nil
}

func assertValidateArchiveRejects(t *testing.T, path string, manifest backup.Manifest) {
	t.Helper()
	if err := backup.ValidateArchive(path, manifest); !errors.Is(err, backup.ErrArtifactFailed) {
		t.Fatalf("ValidateArchive() err = %v, want ErrArtifactFailed", err)
	}
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		t.Fatalf("artifact path state = %v err=%v", info, err)
	}
}

func TestValidateArchiveRejectsAbsoluteMemberPath(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: "/state/vyntrio.db", data: memberData(backup.StateDBMember, manifest)},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsTraversalMember(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: "../outside", data: []byte("x")},
		{name: backup.StateDBMember, data: memberData(backup.StateDBMember, manifest)},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsDuplicateMemberNames(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	db := memberData(backup.StateDBMember, manifest)
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, data: db},
		{name: backup.StateDBMember, data: db},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsSymlinkMember(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, typeflag: tar.TypeSymlink, linkname: "elsewhere"},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsHardLinkMember(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, typeflag: tar.TypeLink, linkname: backup.ConfigMember},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsDirectoryMember(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, typeflag: tar.TypeDir},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsFIFOMember(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, typeflag: tar.TypeFifo},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsUnexpectedAllowlistMember(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, data: memberData(backup.StateDBMember, manifest)},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: "state/extra.db", data: []byte("extra")},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsDigestMismatch(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, data: []byte("tampered-db")},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifest)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	assertValidateArchiveRejects(t, path, manifest)
}

func TestValidateArchiveRejectsManifestMemberMismatch(t *testing.T) {
	manifestInTar := validManifest(t)
	manifestInTar.Members = append(manifestInTar.Members, backup.ManifestMember{
		Name:   backup.StateJournalMem,
		Size:   1,
		SHA256: strings.Repeat("a", 64),
	})
	manifestPassed := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, data: memberData(backup.StateDBMember, manifestPassed)},
		{name: backup.ConfigMember, data: memberData(backup.ConfigMember, manifestPassed)},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifestInTar)},
	})
	assertValidateArchiveRejects(t, path, manifestPassed)
}

func TestValidateArchiveMalformedDoesNotEmitSourceContent(t *testing.T) {
	manifest := validManifest(t)
	path := filepath.Join(t.TempDir(), "bad.tar")
	writeMalformedTar(t, path, []tarEntry{
		{name: backup.StateDBMember, data: []byte("secret-db-content")},
		{name: backup.ConfigMember, data: []byte("secret-config-content")},
		{name: backup.ManifestFileName, data: manifestPayload(t, manifest)},
	})
	err := backup.ValidateArchive(path, manifest)
	if err == nil {
		t.Fatal("expected validation failure")
	}
	msg := err.Error()
	for _, forbidden := range []string{"secret-db-content", "secret-config-content"} {
		if strings.Contains(msg, forbidden) {
			t.Fatalf("error leaked %q: %s", forbidden, msg)
		}
	}
}
