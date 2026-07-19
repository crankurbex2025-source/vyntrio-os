package restore

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

// Mode selects restore CLI behavior.
type Mode int

const (
	ModeValidate Mode = iota
	ModeDryRun
	ModeRestore
)

// Options configures restore paths and ownership overrides for tests.
type Options struct {
	DestinationDir string
	StateRoot      string
	ConfigPath     string
	PreserveRoot   string
	Ownership      Ownership
	ServiceTimeout time.Duration
}

func (o Options) normalized() Options {
	out := o
	if out.DestinationDir == "" {
		out.DestinationDir = DestinationDir
	}
	if out.StateRoot == "" {
		out.StateRoot = StateRoot
	}
	if out.ConfigPath == "" {
		out.ConfigPath = ConfigPath
	}
	if out.ServiceTimeout == 0 {
		out.ServiceTimeout = defaultServiceTimeout
	}
	return out
}

func (o Options) preserveRoot() string {
	if o.PreserveRoot != "" {
		return o.PreserveRoot
	}
	return o.normalized().DestinationDir
}

func (o Options) hostPathForMember(member string) (string, bool) {
	switch member {
	case backup.StateDBMember:
		return o.normalized().StateRoot + "/vyntrio.db", true
	case backup.StateJournalMem:
		return o.normalized().StateRoot + "/vyntrio.db-journal", true
	case backup.StateWALMember:
		return o.normalized().StateRoot + "/vyntrio.db-wal", true
	case backup.StateSHMMember:
		return o.normalized().StateRoot + "/vyntrio.db-shm", true
	case backup.ConfigMember:
		return o.normalized().ConfigPath, true
	default:
		return "", false
	}
}

// Result summarizes a validate/dry-run/restore outcome without sensitive data.
type Result struct {
	ArtifactBasename  string
	MemberCount       int
	Scope             string
	PreserveBasename  string
	RollbackAttempted bool
	RollbackSucceeded bool
}

// Runner executes validate, dry-run, and destructive restore flows.
type Runner struct {
	RootChecker backup.RootChecker
	Service     backup.ServiceController
	Health      backup.HealthProber
	Clock       backup.Clock
	Options     Options
	Stdout      io.Writer
}

// Run executes the selected restore mode.
func (r *Runner) Run(ctx context.Context, mode Mode, artifactBasename string, force bool) (Result, error) {
	if r.RootChecker == nil || !r.RootChecker.IsRoot() {
		return Result{}, ErrNotRoot
	}
	opts := r.Options.normalized()
	artifactPath, err := ResolveArtifactPath(artifactBasename, opts.DestinationDir)
	if err != nil {
		return Result{}, err
	}

	manifest, err := PreflightArchive(artifactPath)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		ArtifactBasename: artifactBasename,
		MemberCount:      len(manifest.Members),
		Scope:            ScopeSummary(manifest),
	}

	switch mode {
	case ModeValidate:
		return result, nil
	case ModeDryRun:
		return result, nil
	case ModeRestore:
		if !force {
			return result, ErrForceRequired
		}
	default:
		return Result{}, ErrUsage
	}

	payloads, err := ExtractRestorableMembers(artifactPath, manifest)
	if err != nil {
		return Result{}, err
	}

	stopped := false
	preserveDir := ""
	restoreMembers := placementTargets(payloads)
	defer func() {
		if stopped {
			_ = r.recoverService(ctx)
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

	now := r.clock().Now().UTC()
	preserveDir, err = createPreservationCopy(opts, now, restoreMembers)
	if err != nil {
		return Result{}, err
	}
	result.PreserveBasename = preserveDirBasename(preserveDir)

	targets, err := placePayloads(opts, payloads)
	if err != nil {
		_ = rollbackFromPreserve(preserveDir, opts, restoreMembers)
		return Result{}, err
	}

	if err := applyOwnership(targets, opts.Ownership); err != nil {
		_ = rollbackFromPreserve(preserveDir, opts, restoreMembers)
		return Result{}, fmt.Errorf("%w: %v", ErrOwnershipFailed, err)
	}

	stopped = false
	if err := r.completeRestore(ctx); err != nil {
		fc := postPlacementContext{
			result:      result,
			preserveDir: preserveDir,
			opts:        opts,
			members:     restoreMembers,
			targets:     targets,
		}
		out, handleErr := r.handlePostPlacementFailure(ctx, fc, err)
		return out, handleErr
	}

	return result, nil
}

func (r *Runner) recoverService(ctx context.Context) error {
	if err := r.Service.Start(ctx); err != nil {
		return err
	}
	active, err := r.Service.IsActive(ctx)
	return errIfNotActive(active, err)
}

func errIfNotActive(active bool, err error) error {
	if err != nil || !active {
		return ErrServiceStartFailed
	}
	return nil
}

func (r *Runner) clock() backup.Clock {
	if r.Clock != nil {
		return r.Clock
	}
	return backup.RealClock()
}

func preserveDirBasename(path string) string {
	if path == "" {
		return ""
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
