package restore_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/restore"
)

type fakeRoot struct{ root bool }

func (f fakeRoot) IsRoot() bool { return f.root }

type fakeService struct {
	active    bool
	inactive  bool
	stopErr   error
	startErr  error
	calls     []string
}

func (f *fakeService) IsActive(context.Context) (bool, error) {
	f.calls = append(f.calls, "is-active")
	return f.active, nil
}

func (f *fakeService) Stop(context.Context) error {
	f.calls = append(f.calls, "stop")
	if f.stopErr != nil {
		return f.stopErr
	}
	f.active = false
	f.inactive = true
	return nil
}

func (f *fakeService) IsInactive(context.Context) (bool, error) {
	f.calls = append(f.calls, "is-inactive")
	return f.inactive, nil
}

func (f *fakeService) Start(context.Context) error {
	f.calls = append(f.calls, "start")
	if f.startErr != nil {
		return f.startErr
	}
	f.active = true
	f.inactive = false
	return nil
}

type fakeHealth struct {
	err error
}

func (f fakeHealth) Probe(context.Context) error { return f.err }

type sequentialHealth struct {
	errs []error
	n    int
}

func (f *sequentialHealth) Probe(context.Context) error {
	if f.n < len(f.errs) {
		err := f.errs[f.n]
		f.n++
		return err
	}
	return nil
}

type stopAfterFirstService struct {
	inner      *fakeService
	stopCalls  int
	failSecond bool
}

func (s *stopAfterFirstService) IsActive(ctx context.Context) (bool, error) {
	return s.inner.IsActive(ctx)
}

func (s *stopAfterFirstService) Stop(ctx context.Context) error {
	s.stopCalls++
	if s.failSecond && s.stopCalls > 1 {
		return errors.New("stop failed during rollback")
	}
	return s.inner.Stop(ctx)
}

func (s *stopAfterFirstService) IsInactive(ctx context.Context) (bool, error) {
	return s.inner.IsInactive(ctx)
}

func (s *stopAfterFirstService) Start(ctx context.Context) error {
	return s.inner.Start(ctx)
}

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func testLayout(t *testing.T) (destination, stateRoot, configPath, preserveRoot string) {
	t.Helper()
	root := t.TempDir()
	destination = filepath.Join(root, "backups")
	stateRoot = filepath.Join(root, "state")
	configPath = filepath.Join(root, "config.toml")
	preserveRoot = destination
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(stateRoot, "vyntrio.db"), []byte("live-db"))
	writeFile(t, configPath, []byte("live-config"))
	return destination, stateRoot, configPath, preserveRoot
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
}

func buildArtifact(t *testing.T, destination string, dbData, cfgData []byte, meta backup.ReleaseMetadata) (string, backup.Manifest) {
	t.Helper()
	createdAt := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	timestamp := createdAt.Format("20060102T150405.000000000Z")
	basename := backup.FormatVersion + "_" + timestamp + ".tar"
	path := filepath.Join(destination, basename)
	sources := []backup.SourceMember{
		{ArchiveName: backup.StateDBMember, SourcePath: writeTempSource(t, dbData)},
		{ArchiveName: backup.ConfigMember, SourcePath: writeTempSource(t, cfgData)},
	}
	manifest, err := backup.BuildArchive(path, createdAt, meta, sources)
	if err != nil {
		t.Fatal(err)
	}
	return basename, manifest
}

func writeTempSource(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source")
	writeFile(t, path, data)
	return path
}

func testRunner(t *testing.T, destination, stateRoot, configPath, preserveRoot string) restore.Runner {
	t.Helper()
	return restore.Runner{
		RootChecker: fakeRoot{root: true},
		Service:     &fakeService{active: true},
		Health:      fakeHealth{},
		Clock:       fixedClock{now: time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC)},
		Options: restore.Options{
			DestinationDir: destination,
			StateRoot:      stateRoot,
			ConfigPath:     configPath,
			PreserveRoot:   preserveRoot,
			Ownership: restore.Ownership{
				StateUID:  os.Getuid(),
				StateGID:  os.Getgid(),
				ConfigUID: os.Getuid(),
				ConfigGID: os.Getgid(),
			},
		},
	}
}

func TestResolveArtifactPathRejectsTraversalAndTmp(t *testing.T) {
	destination := t.TempDir()
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	cases := []string{
		"",
		"../escape.tar",
		"vyntrio-backup-v1_20260715T100000.000000000Z.tar.tmp",
		"not-a-backup.tar",
	}
	for _, name := range cases {
		if _, err := restore.ResolveArtifactPath(name, destination); !errors.Is(err, restore.ErrArtifactInvalid) {
			t.Fatalf("expected invalid artifact for %q, got %v", name, err)
		}
	}
}

func TestPreflightArchiveRejectsDigestMismatch(t *testing.T) {
	destination, _, _, _ := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("db"), []byte("cfg"), backup.ReleaseMetadata{})
	path := filepath.Join(destination, basename)
	if err := os.WriteFile(path, []byte("not-a-tar"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := restore.PreflightArchive(path); !errors.Is(err, restore.ErrArtifactInvalid) {
		t.Fatalf("expected preflight failure, got %v", err)
	}
}

func TestValidateModeDoesNotMutateLiveState(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), backup.ReleaseMetadata{})
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	result, err := runner.Run(context.Background(), restore.ModeValidate, basename, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.MemberCount != 2 {
		t.Fatalf("expected 2 members, got %d", result.MemberCount)
	}
	assertFileContents(t, filepath.Join(stateRoot, "vyntrio.db"), []byte("live-db"))
	assertFileContents(t, configPath, []byte("live-config"))
}

func TestRestoreRequiresForce(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), backup.ReleaseMetadata{})
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	_, err := runner.Run(context.Background(), restore.ModeRestore, basename, false)
	if !errors.Is(err, restore.ErrForceRequired) {
		t.Fatalf("expected force required, got %v", err)
	}
	assertFileContents(t, filepath.Join(stateRoot, "vyntrio.db"), []byte("live-db"))
}

func TestRestoreRejectsDirtyMigrationManifest(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	meta := backup.ReleaseMetadata{
		MigrationVersion:    "3",
		MigrationDirtyKnown: true,
		MigrationDirty:      true,
	}
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), meta)
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	_, err := runner.Run(context.Background(), restore.ModeValidate, basename, false)
	if !errors.Is(err, restore.ErrCompatibility) {
		t.Fatalf("expected compatibility failure, got %v", err)
	}
}

func TestRestoreForceReplacesStateAndConfigWithPreserveCopy(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), backup.ReleaseMetadata{})
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	result, err := runner.Run(context.Background(), restore.ModeRestore, basename, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.PreserveBasename == "" || !strings.HasPrefix(result.PreserveBasename, "preserve-v1_") {
		t.Fatalf("unexpected preserve basename: %q", result.PreserveBasename)
	}
	assertFileContents(t, filepath.Join(stateRoot, "vyntrio.db"), []byte("archive-db"))
	assertFileContents(t, configPath, []byte("archive-cfg"))
	preservePath := filepath.Join(preserveRoot, result.PreserveBasename, backup.StateDBMember)
	assertFileContents(t, preservePath, []byte("live-db"))
}

func TestDryRunDoesNotMutateLiveState(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), backup.ReleaseMetadata{})
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	_, err := runner.Run(context.Background(), restore.ModeDryRun, basename, false)
	if err != nil {
		t.Fatal(err)
	}
	assertFileContents(t, filepath.Join(stateRoot, "vyntrio.db"), []byte("live-db"))
}

func TestPostRestoreHealthFailureRollsBackLiveState(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), backup.ReleaseMetadata{})
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	runner.Health = &sequentialHealth{errs: []error{restore.ErrHealthProbeFailed}}
	result, err := runner.Run(context.Background(), restore.ModeRestore, basename, true)
	if !errors.Is(err, restore.ErrPostRestoreRollbackSucceeded) {
		t.Fatalf("expected rollback-succeeded outcome, got %v", err)
	}
	if !result.RollbackAttempted || !result.RollbackSucceeded {
		t.Fatalf("expected rollback flags set, got %+v", result)
	}
	assertFileContents(t, filepath.Join(stateRoot, "vyntrio.db"), []byte("live-db"))
	assertFileContents(t, configPath, []byte("live-config"))
}

func TestPostRestoreHealthFailureRollbackStopFailure(t *testing.T) {
	destination, stateRoot, configPath, preserveRoot := testLayout(t)
	basename, _ := buildArtifact(t, destination, []byte("archive-db"), []byte("archive-cfg"), backup.ReleaseMetadata{})
	inner := &fakeService{active: true}
	runner := testRunner(t, destination, stateRoot, configPath, preserveRoot)
	runner.Service = &stopAfterFirstService{inner: inner, failSecond: true}
	runner.Health = &sequentialHealth{errs: []error{restore.ErrReadinessProbeFailed}}
	result, err := runner.Run(context.Background(), restore.ModeRestore, basename, true)
	if !errors.Is(err, restore.ErrPostRestoreRollbackFailed) {
		t.Fatalf("expected rollback-failed outcome, got %v", err)
	}
	if !result.RollbackAttempted || result.RollbackSucceeded {
		t.Fatalf("expected rollback attempted without success, got %+v", result)
	}
}

func assertFileContents(t *testing.T, path string, want []byte) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(want) {
		t.Fatalf("file %s = %q, want %q", path, data, want)
	}
}
