package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

// Options configures a backup run.
type Options struct {
	StateRoot      string
	ConfigPath     string
	DestinationDir string
	ServiceTimeout time.Duration
	HTTPTimeout    time.Duration
}

// Result summarizes a successful backup without sensitive content.
type Result struct {
	ArtifactName string
	MemberCount  int
}

// Runner executes the backup workflow.
type Runner struct {
	RootChecker        RootChecker
	Service            ServiceController
	Health             HealthProber
	VersionFetcher     VersionFetcher
	MigrationReader    MigrationReader
	Clock              Clock
	StatusPublisher    RunStatusPublisher
	PrepareDestination func(string) error
	Options            Options
	Stdout             io.Writer
}

// Run executes the approved backup contract.
func (r *Runner) Run(ctx context.Context) (Result, error) {
	if r.RootChecker == nil || !r.RootChecker.IsRoot() {
		return Result{}, ErrNotRoot
	}
	opts := r.normalizedOptions()
	clock := r.clock()

	active, err := r.Service.IsActive(ctx)
	if err != nil || !active {
		return Result{}, ErrServiceNotActive
	}

	releaseMeta, _ := r.fetchReleaseMetadata(ctx)

	stopped := false
	attempted := false
	publishFailure := ""
	defer func() {
		if stopped {
			if recoveryClass := r.recoverService(ctx, opts); recoveryClass != "" {
				publishFailure = recoveryClass
			}
		}
		if attempted && publishFailure != "" {
			r.recordAttemptedFailure(ctx, opts.StateRoot, r.clock().Now(), publishFailure)
		}
	}()

	if err := r.Service.Stop(ctx); err != nil {
		return Result{}, ErrServiceStopFailed
	}
	stopped = true

	inactive, err := r.Service.IsInactive(ctx)
	if err != nil || !inactive {
		return Result{}, ErrServiceInactiveUnknown
	}
	attempted = true

	sources, err := SelectSourceMembers(opts.StateRoot, opts.ConfigPath)
	if err != nil {
		publishFailure = FailureClassFromError(err)
		return Result{}, err
	}

	migrationMeta, err := r.MigrationReader.Read(ctx, sourcePathForMember(sources, StateDBMember))
	if err == nil {
		releaseMeta = mergeMetadata(releaseMeta, migrationMeta)
	}

	if err := r.prepareDestination(opts.DestinationDir); err != nil {
		publishFailure = FailureClassFromError(err)
		return Result{}, err
	}

	createdAt := clock.Now().UTC()
	timestamp := createdAt.Format("20060102T150405.000000000Z")
	finalPath := artifactPath(opts.DestinationDir, timestamp)
	tempPath := tempArtifactPath(opts.DestinationDir, timestamp)

	if _, err := os.Lstat(finalPath); err == nil {
		publishFailure = backupstatus.FailureArtifact
		return Result{}, ErrArtifactCollision
	} else if !os.IsNotExist(err) {
		publishFailure = backupstatus.FailureArtifact
		return Result{}, fmt.Errorf("%w: artifact path inaccessible", ErrArtifactFailed)
	}

	manifest, err := BuildArchive(tempPath, createdAt, releaseMeta, sources)
	if err != nil {
		_ = os.Remove(tempPath)
		publishFailure = FailureClassFromError(err)
		return Result{}, err
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		publishFailure = backupstatus.FailureArtifact
		return Result{}, fmt.Errorf("%w: publish artifact", ErrArtifactFailed)
	}
	if err := finalizeArtifactMode(finalPath); err != nil {
		publishFailure = backupstatus.FailureArtifact
		return Result{}, fmt.Errorf("%w: set artifact permissions", ErrArtifactFailed)
	}

	if err := r.Service.Start(ctx); err != nil {
		publishFailure = backupstatus.FailureRestart
		r.writeFailure(FailureRestart, finalPath)
		return Result{}, ErrServiceStartFailed
	}
	stopped = false

	active, err = r.Service.IsActive(ctx)
	if err != nil || !active {
		publishFailure = backupstatus.FailureRestart
		r.writeFailure(FailureRestart, finalPath)
		return Result{}, ErrServiceStartFailed
	}
	if err := r.Health.Probe(ctx); err != nil {
		category := FailureHealth
		publishFailure = backupstatus.FailureHealth
		exitErr := ErrHealthProbeFailed
		if errors.Is(err, ErrReadinessProbeFailed) {
			category = FailureReadiness
			publishFailure = backupstatus.FailureReadiness
			exitErr = ErrReadinessProbeFailed
		}
		r.writeFailure(category, finalPath)
		return Result{}, exitErr
	}

	completedAt := clock.Now().UTC()
	if err := r.publishSucceeded(ctx, opts.StateRoot, completedAt); err != nil {
		return Result{}, ErrStatusPublishFailed
	}

	r.writeSuccess(finalPath, len(manifest.Members))
	return Result{
		ArtifactName: fmt.Sprintf("%s_%s.tar", FormatVersion, timestamp),
		MemberCount:  len(manifest.Members) + 1,
	}, nil
}

func (r *Runner) recoverService(ctx context.Context, opts Options) string {
	if err := r.Service.Start(ctx); err != nil {
		r.writeFailure(FailureRestart, "")
		return backupstatus.FailureRestart
	}
	active, err := r.Service.IsActive(ctx)
	if err != nil || !active {
		r.writeFailure(FailureRestart, "")
		return backupstatus.FailureRestart
	}
	if err := r.Health.Probe(ctx); err != nil {
		if errors.Is(err, ErrReadinessProbeFailed) {
			r.writeFailure(FailureReadiness, "")
			return backupstatus.FailureReadiness
		}
		r.writeFailure(FailureHealth, "")
		return backupstatus.FailureHealth
	}
	return ""
}

func (r *Runner) recordAttemptedFailure(ctx context.Context, stateRoot string, completedAt time.Time, failureClass string) {
	publisher := r.statusPublisher()
	if publisher == nil {
		return
	}
	_ = publisher.PublishFailed(ctx, stateRoot, completedAt, failureClass)
}

func (r *Runner) publishSucceeded(ctx context.Context, stateRoot string, completedAt time.Time) error {
	publisher := r.statusPublisher()
	if publisher == nil {
		return nil
	}
	return publisher.PublishSucceeded(ctx, stateRoot, completedAt)
}

func (r *Runner) statusPublisher() RunStatusPublisher {
	if r.StatusPublisher != nil {
		return r.StatusPublisher
	}
	return DefaultRunStatusPublisher()
}

func (r *Runner) fetchReleaseMetadata(ctx context.Context) (ReleaseMetadata, error) {
	if r.VersionFetcher == nil {
		return ReleaseMetadata{}, nil
	}
	return r.VersionFetcher.Fetch(ctx)
}

func (r *Runner) writeSuccess(finalPath string, memberCount int) {
	if r.Stdout == nil {
		return
	}
	_, _ = fmt.Fprintf(r.Stdout, "backup succeeded: artifact=%s members=%d\n", basename(finalPath), memberCount)
}

func (r *Runner) writeFailure(category FailureCategory, finalPath string) {
	if r.Stdout == nil {
		return
	}
	if finalPath == "" {
		_, _ = fmt.Fprintf(r.Stdout, "backup failed: category=%s\n", category)
		return
	}
	_, _ = fmt.Fprintf(r.Stdout, "backup failed: category=%s artifact=%s\n", category, basename(finalPath))
}

func (r *Runner) normalizedOptions() Options {
	opts := r.Options
	if opts.StateRoot == "" {
		opts.StateRoot = StateRoot
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = ConfigPath
	}
	if opts.DestinationDir == "" {
		opts.DestinationDir = DestinationDir
	}
	if opts.ServiceTimeout == 0 {
		opts.ServiceTimeout = defaultServiceTimeout
	}
	if opts.HTTPTimeout == 0 {
		opts.HTTPTimeout = defaultHTTPTimeout
	}
	return opts
}

func (r *Runner) clock() Clock {
	if r.Clock == nil {
		return RealClock()
	}
	return r.Clock
}

func (r *Runner) prepareDestination(path string) error {
	if r.PrepareDestination != nil {
		return r.PrepareDestination(path)
	}
	return EnsureDestination(path)
}

func mergeMetadata(base, extra ReleaseMetadata) ReleaseMetadata {
	if extra.APIVersion != "" {
		base.APIVersion = extra.APIVersion
	}
	if extra.APICommit != "" {
		base.APICommit = extra.APICommit
	}
	if extra.MigrationVersion != "" {
		base.MigrationVersion = extra.MigrationVersion
		base.MigrationDirty = extra.MigrationDirty
		base.MigrationDirtyKnown = extra.MigrationDirtyKnown
	}
	return base
}

func sourcePathForMember(sources []SourceMember, name string) string {
	for _, source := range sources {
		if source.ArchiveName == name {
			return source.SourcePath
		}
	}
	return ""
}

func basename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// HTTPVersionFetcher loads /api/v1/version while the service is active.
type HTTPVersionFetcher struct {
	Client *http.Client
	URL    string
}

func (f HTTPVersionFetcher) Fetch(ctx context.Context) (ReleaseMetadata, error) {
	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	url := f.URL
	if url == "" {
		url = VersionURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ReleaseMetadata{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return ReleaseMetadata{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return ReleaseMetadata{}, fmt.Errorf("version endpoint status %d", resp.StatusCode)
	}
	var payload struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ReleaseMetadata{}, err
	}
	return ReleaseMetadata{
		APIVersion: payload.Version,
		APICommit:  payload.Commit,
	}, nil
}
