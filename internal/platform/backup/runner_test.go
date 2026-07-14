package backup_test

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

type fakeRoot struct{ root bool }

func (f fakeRoot) IsRoot() bool { return f.root }

type fakeService struct {
	active        bool
	inactive      bool
	stopErr       error
	startErr      error
	isActiveErr   error
	isInactiveErr error
	calls         []string
}

func (f *fakeService) IsActive(context.Context) (bool, error) {
	f.calls = append(f.calls, "is-active")
	if f.isActiveErr != nil {
		return false, f.isActiveErr
	}
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
	if f.isInactiveErr != nil {
		return false, f.isInactiveErr
	}
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

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func fastProbePolicy(deadline time.Duration) backup.PostStartProbePolicy {
	current := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	return backup.PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      deadline,
		Now:           func() time.Time { return current },
		Sleep: func(ctx context.Context, d time.Duration) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			current = current.Add(d)
			return nil
		},
	}
}

func loopbackProber(t *testing.T, srvURL string, policy backup.PostStartProbePolicy) backup.LoopbackHealthProber {
	t.Helper()
	return backup.LoopbackHealthProber{BaseURL: srvURL, Policy: policy}
}

func testPrepareDestination(path string) error {
	return os.MkdirAll(path, 0o700)
}

func writeRegular(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o640); err != nil {
		t.Fatal(err)
	}
}

func setupState(t *testing.T) (stateRoot, configPath, dest string) {
	t.Helper()
	root := t.TempDir()
	stateRoot = filepath.Join(root, "state")
	dest = filepath.Join(root, "backups")
	configPath = filepath.Join(root, "config.toml")
	writeRegular(t, filepath.Join(stateRoot, backup.MainDBName), []byte("db-bytes"))
	writeRegular(t, configPath, []byte("bind_address=\"127.0.0.1\"\n"))
	return stateRoot, configPath, dest
}

func newTestRunner(t *testing.T, svc *fakeService, health backup.HealthProber, stateRoot, configPath, dest string) backup.Runner {
	t.Helper()
	return backup.Runner{
		RootChecker:        fakeRoot{root: true},
		Service:            svc,
		Health:             health,
		VersionFetcher:     nil,
		MigrationReader:    backup.SQLMigrationReader{},
		Clock:              fixedClock{now: time.Date(2026, 7, 13, 15, 4, 5, 0, time.UTC)},
		PrepareDestination: testPrepareDestination,
		Stdout:             io.Discard,
		Options: backup.Options{
			StateRoot:      stateRoot,
			ConfigPath:     configPath,
			DestinationDir: dest,
		},
	}
}

func TestRunRejectsNonRootBeforeServiceActions(t *testing.T) {
	svc := &fakeService{active: true}
	runner := newTestRunner(t, svc, fakeHealth{}, t.TempDir(), filepath.Join(t.TempDir(), "config.toml"), t.TempDir())
	runner.RootChecker = fakeRoot{root: false}
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrNotRoot) {
		t.Fatalf("err = %v, want ErrNotRoot", err)
	}
	if len(svc.calls) != 0 {
		t.Fatalf("service calls = %v, want none", svc.calls)
	}
}

func TestRunRequiresActiveServiceBeforeStop(t *testing.T) {
	svc := &fakeService{active: false}
	stateRoot, configPath, dest := setupState(t)
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrServiceNotActive) {
		t.Fatalf("err = %v", err)
	}
	if len(svc.calls) != 1 || svc.calls[0] != "is-active" {
		t.Fatalf("calls = %v", svc.calls)
	}
}

func TestRunStopThenInactiveBeforeReadOrder(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if svc.calls[0] != "is-active" || svc.calls[1] != "stop" || svc.calls[2] != "is-inactive" {
		t.Fatalf("calls = %v", svc.calls)
	}
}

func TestRunRejectsMissingMainDB(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	root := t.TempDir()
	stateRoot := filepath.Join(root, "state")
	dest := filepath.Join(root, "backups")
	configPath := filepath.Join(root, "config.toml")
	writeRegular(t, configPath, []byte("config"))
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrSourceInvalid) {
		t.Fatalf("err = %v", err)
	}
}

func TestRunRejectsMissingRequiredConfig(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	root := t.TempDir()
	stateRoot := filepath.Join(root, "state")
	dest := filepath.Join(root, "backups")
	configPath := filepath.Join(root, "config.toml")
	writeRegular(t, filepath.Join(stateRoot, backup.MainDBName), []byte("db-bytes"))
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	runner.Stdout = &buf
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrSourceInvalid) {
		t.Fatalf("err = %v, want ErrSourceInvalid", err)
	}
	if entries, readErr := os.ReadDir(dest); readErr != nil {
		if !os.IsNotExist(readErr) {
			t.Fatalf("ReadDir(dest) err = %v", readErr)
		}
	} else if len(entries) != 0 {
		t.Fatalf("dest entries = %v, want none", entries)
	}
	if !containsCall(svc.calls, "stop") || !containsCall(svc.calls, "start") {
		t.Fatalf("calls = %v, want stop then restart attempt after failure", svc.calls)
	}
	out := buf.String()
	for _, forbidden := range []string{"db-bytes", configPath, stateRoot} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("stdout leaked %q in %q", forbidden, out)
		}
	}
}

func TestRunRejectsSymlinkSource(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	target := filepath.Join(t.TempDir(), "secret")
	writeRegular(t, target, []byte("x"))
	_ = os.Remove(filepath.Join(stateRoot, backup.MainDBName))
	if err := os.Symlink(target, filepath.Join(stateRoot, backup.MainDBName)); err != nil {
		t.Fatal(err)
	}
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrSourceInvalid) {
		t.Fatalf("err = %v", err)
	}
}

func TestRunIncludesOnlyExistingSidecars(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	writeRegular(t, filepath.Join(stateRoot, backup.JournalSidecar), []byte("journal"))
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	artifact := filepath.Join(dest, result.ArtifactName)
	manifest, members := readArchive(t, artifact)
	if _, ok := members[backup.StateJournalMem]; !ok {
		t.Fatal("expected journal sidecar in archive")
	}
	if _, ok := members[backup.StateWALMember]; ok {
		t.Fatal("unexpected wal member")
	}
	if manifest.FormatVersion != backup.FormatVersion {
		t.Fatalf("format = %q", manifest.FormatVersion)
	}
}

func TestRunSuccessOrderAndHealthProbes(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	wantCalls := []string{"is-active", "stop", "is-inactive", "start", "is-active"}
	if len(svc.calls) < len(wantCalls) {
		t.Fatalf("calls = %v", svc.calls)
	}
	for i, call := range wantCalls {
		if svc.calls[i] != call {
			t.Fatalf("calls = %v, want prefix %v", svc.calls, wantCalls)
		}
	}
	if !strings.HasPrefix(result.ArtifactName, backup.FormatVersion+"_") {
		t.Fatalf("artifact = %q", result.ArtifactName)
	}
}

func TestRunTamperedArchiveFailsValidation(t *testing.T) {
	createdAt := time.Date(2026, 7, 13, 15, 4, 5, 0, time.UTC)
	sources := []backup.SourceMember{
		{ArchiveName: backup.StateDBMember, SourcePath: writeTempFile(t, []byte("db"))},
		{ArchiveName: backup.ConfigMember, SourcePath: writeTempFile(t, []byte("cfg"))},
	}
	temp := filepath.Join(t.TempDir(), "bad.tar.tmp")
	manifest, err := backup.BuildArchive(temp, createdAt, backup.ReleaseMetadata{}, sources)
	if err != nil {
		t.Fatalf("BuildArchive() error: %v", err)
	}
	f, err := os.OpenFile(temp, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteAt([]byte("tamper"), 200); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if err := backup.ValidateArchive(temp, manifest); err == nil {
		t.Fatal("expected validation failure")
	}
}

func TestRunAtomicPublicationRefusesCollision(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	timestamp := "20260713T150405.000000000Z"
	final := filepath.Join(dest, backup.FormatVersion+"_"+timestamp+".tar")
	if err := os.MkdirAll(dest, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(final, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	runner.Clock = fixedClock{now: time.Date(2026, 7, 13, 15, 4, 5, 0, time.UTC)}
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrArtifactCollision) {
		t.Fatalf("err = %v", err)
	}
}

func TestRunRestartFailurePreservesArtifact(t *testing.T) {
	svc := &fakeService{active: true, inactive: true, startErr: errors.New("start failed")}
	stateRoot, configPath, dest := setupState(t)
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrServiceStartFailed) {
		t.Fatalf("err = %v", err)
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.HasSuffix(entries[0].Name(), ".tar") {
		t.Fatalf("entries = %v", entries)
	}
}

func TestRunHealthFailurePreservesArtifact(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	runner := newTestRunner(t, svc, fakeHealth{err: backup.ErrHealthProbeFailed}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrHealthProbeFailed) {
		t.Fatalf("err = %v", err)
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 1 {
		t.Fatalf("entries = %v", entries)
	}
}

func TestRunReadinessFailureAfterPublication(t *testing.T) {
	readyCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		readyCalls++
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, loopbackProber(t, srv.URL, fastProbePolicy(5*time.Millisecond)), stateRoot, configPath, dest)
	runner.Stdout = &buf
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrReadinessProbeFailed) {
		t.Fatalf("err = %v, want ErrReadinessProbeFailed", err)
	}
	if readyCalls < 2 {
		t.Fatalf("ready calls = %d, want bounded retries", readyCalls)
	}
	if startCount := countCalls(svc.calls, "start"); startCount != 1 {
		t.Fatalf("start calls = %d, want exactly one", startCount)
	}
	out := buf.String()
	if strings.Contains(out, "backup succeeded") {
		t.Fatalf("stdout must not report success: %q", out)
	}
	if !strings.Contains(out, string(backup.FailureReadiness)) {
		t.Fatalf("stdout = %q, want readiness failure category", out)
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.HasSuffix(entries[0].Name(), ".tar") {
		t.Fatalf("entries = %v, want preserved published artifact", entries)
	}
	for _, forbidden := range []string{"db-bytes", "bind_address", configPath, stateRoot} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("stdout leaked %q in %q", forbidden, out)
		}
	}
}

func TestRunTransientHealthStartupRaceSucceeds(t *testing.T) {
	healthCalls := 0
	callLog := []string{}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		healthCalls++
		callLog = append(callLog, "/healthz")
		if healthCalls < 3 {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		callLog = append(callLog, "/readyz")
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, loopbackProber(t, srv.URL, fastProbePolicy(5*time.Second)), stateRoot, configPath, dest)
	runner.Stdout = &buf
	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.Contains(buf.String(), "backup succeeded") {
		t.Fatalf("stdout = %q, want success", buf.String())
	}
	if healthCalls != 3 {
		t.Fatalf("health calls = %d, want 3", healthCalls)
	}
	if countCalls(callLog, "/readyz") != 1 {
		t.Fatalf("probe order = %v", callLog)
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %v, want one artifact", entries)
	}
	if result.ArtifactName != entries[0].Name() {
		t.Fatalf("artifact = %q, dir entry = %q", result.ArtifactName, entries[0].Name())
	}
}

func TestRunTransientReadinessStartupRaceSucceeds(t *testing.T) {
	readyCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		readyCalls++
		if readyCalls < 3 {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, loopbackProber(t, srv.URL, fastProbePolicy(5*time.Second)), stateRoot, configPath, dest)
	runner.Stdout = &buf
	_, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.Contains(buf.String(), "backup succeeded") {
		t.Fatalf("stdout = %q, want success", buf.String())
	}
	if readyCalls != 3 {
		t.Fatalf("ready calls = %d, want 3", readyCalls)
	}
}

func TestRunPersistentHealthFailureAfterPublication(t *testing.T) {
	healthCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		healthCalls++
		http.Error(w, "down", http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("readyz must not be probed before healthz succeeds")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, loopbackProber(t, srv.URL, fastProbePolicy(5*time.Millisecond)), stateRoot, configPath, dest)
	runner.Stdout = &buf
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrHealthProbeFailed) {
		t.Fatalf("err = %v, want ErrHealthProbeFailed", err)
	}
	if healthCalls < 2 {
		t.Fatalf("health calls = %d, want bounded retries", healthCalls)
	}
	if startCount := countCalls(svc.calls, "start"); startCount != 1 {
		t.Fatalf("start calls = %d, want exactly one", startCount)
	}
	if strings.Contains(buf.String(), "backup succeeded") {
		t.Fatalf("stdout must not report success")
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %v, want preserved artifact", entries)
	}
}

func TestRunProbeContextCancellationDuringRetry(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner := newTestRunner(t, svc, loopbackProber(t, srv.URL, fastProbePolicy(time.Minute)), stateRoot, configPath, dest)
	_, err := runner.Run(ctx)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !errors.Is(err, context.Canceled) && !errors.Is(err, backup.ErrHealthProbeFailed) {
		t.Fatalf("err = %v, want cancellation or health probe failure", err)
	}
}

func TestRunProbeRetryDoesNotLeakSensitiveContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "secret-db-content", http.StatusServiceUnavailable)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, loopbackProber(t, srv.URL, fastProbePolicy(2*time.Millisecond)), stateRoot, configPath, dest)
	runner.Stdout = &buf
	_, err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected failure")
	}
	out := buf.String()
	for _, forbidden := range []string{"secret-db-content", "db-bytes", "bind_address", configPath, stateRoot} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("stdout leaked %q in %q", forbidden, out)
		}
	}
}

func countCalls(calls []string, want string) int {
	n := 0
	for _, call := range calls {
		if call == want {
			n++
		}
	}
	return n
}

func TestRunDestinationFailureDoesNotCreateArtifact(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	if err := os.Symlink(t.TempDir(), dest); err != nil {
		t.Fatal(err)
	}
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	runner.PrepareDestination = nil
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrDestinationUnsafe) {
		t.Fatalf("err = %v, want ErrDestinationUnsafe", err)
	}
	parent := filepath.Dir(dest)
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar") || strings.HasSuffix(entry.Name(), ".tar.tmp") {
			t.Fatalf("unexpected artifact %q after destination failure", entry.Name())
		}
	}
}

func TestValidateArchiveRejectsTraversalAndDuplicates(t *testing.T) {
	if err := backup.ValidateArchiveMemberName("state/../vyntrio.db"); err == nil {
		t.Fatal("expected traversal rejection")
	}
	if err := backup.ValidateArchiveMemberName("/state/vyntrio.db"); err == nil {
		t.Fatal("expected absolute rejection")
	}
}

func TestManifestRoundTripDeterministic(t *testing.T) {
	createdAt := time.Date(2026, 7, 13, 15, 4, 5, 0, time.UTC)
	members := []backup.ManifestMember{{Name: backup.StateDBMember, Size: 3, SHA256: "abc"}}
	manifest := backup.NewManifest(createdAt, backup.ReleaseMetadata{APIVersion: "1"}, members)
	data, err := manifest.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"format_version":"vyntrio-backup-v1"`) {
		t.Fatalf("json = %s", data)
	}
}

func readArchive(t *testing.T, path string) (backup.Manifest, map[string][]byte) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	tr := tar.NewReader(f)
	members := make(map[string][]byte)
	var manifest backup.Manifest
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		members[hdr.Name] = data
		if hdr.Name == backup.ManifestFileName {
			manifest, err = backup.DecodeManifest(data)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return manifest, members
}

func writeTempFile(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "file")
	writeRegular(t, path, data)
	return path
}

func TestDigestMatchesArchiveBytes(t *testing.T) {
	data := []byte("payload")
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) == "" {
		t.Fatal("unexpected empty digest")
	}
}

func TestSelectSourceMembersOptionalSidecars(t *testing.T) {
	stateRoot, configPath, _ := setupState(t)
	members, err := backup.SelectSourceMembers(stateRoot, configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Fatalf("members = %d, want 2", len(members))
	}
}

func TestRunFailureAfterStopAttemptsRestart(t *testing.T) {
	svc := &fakeService{active: true, isInactiveErr: errors.New("inactive unknown")}
	stateRoot, configPath, dest := setupState(t)
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	_, err := runner.Run(context.Background())
	if !errors.Is(err, backup.ErrServiceInactiveUnknown) {
		t.Fatalf("err = %v", err)
	}
	if !containsCall(svc.calls, "start") {
		t.Fatalf("calls = %v, want restart attempt", svc.calls)
	}
}

func containsCall(calls []string, want string) bool {
	for _, call := range calls {
		if call == want {
			return true
		}
	}
	return false
}

func TestRunStdoutDoesNotIncludeSensitiveContent(t *testing.T) {
	svc := &fakeService{active: true, inactive: true}
	stateRoot, configPath, dest := setupState(t)
	var buf bytes.Buffer
	runner := newTestRunner(t, svc, fakeHealth{}, stateRoot, configPath, dest)
	runner.Stdout = &buf
	_, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, forbidden := range []string{"db-bytes", "bind_address", "sha256", configPath, stateRoot} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("stdout leaked %q in %q", forbidden, out)
		}
	}
}

func TestValidateArchiveMemberNameExported(t *testing.T) {
	if err := backup.ValidateArchiveMemberName("state/vyntrio.db"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
