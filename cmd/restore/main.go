// Package main is the root-only offline restore command for Vyntrio OS.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/restore"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	if args[0] == "validate" {
		return runMode(args[1:], restore.ModeValidate, false)
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		usage()
		return 0
	}

	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dryRun := fs.Bool("dry-run", false, "validate artifact and print restore scope without mutation")
	force := fs.Bool("force", false, "required for destructive restore")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	remaining := fs.Args()
	if len(remaining) != 1 {
		usage()
		return 2
	}
	mode := restore.ModeRestore
	if *dryRun {
		mode = restore.ModeDryRun
	}
	return runRestore(remaining[0], mode, *force)
}

func runMode(args []string, mode restore.Mode, force bool) int {
	if len(args) != 1 {
		usage()
		return 2
	}
	return runRestore(args[0], mode, force)
}

func runRestore(artifactBasename string, mode restore.Mode, force bool) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	runner := restore.Runner{
		RootChecker: backup.OSRootChecker{},
		Service:     backup.SystemdController{},
		Health:      backup.LoopbackHealthProber{},
		Clock:       backup.RealClock(),
		Stdout:      os.Stdout,
	}

	result, err := runner.Run(ctx, mode, artifactBasename, force)
	if err != nil {
		category := restore.FailureClassFromError(err)
		switch {
		case errors.Is(err, restore.ErrPostRestoreRollbackSucceeded):
			_, _ = fmt.Fprintf(os.Stderr,
				"restore failed: rollback=succeeded preserve=%s category=%s\n",
				result.PreserveBasename, category)
		case errors.Is(err, restore.ErrPostRestoreRollbackFailed):
			_, _ = fmt.Fprintf(os.Stderr,
				"restore failed: rollback=failed preserve=%s category=%s\n",
				result.PreserveBasename, category)
		default:
			_, _ = fmt.Fprintf(os.Stderr, "restore failed: category=%s\n", category)
		}
		return 1
	}

	switch mode {
	case restore.ModeValidate:
		_, _ = fmt.Fprintf(os.Stdout, "restore validate succeeded: artifact=%s members=%d\n",
			result.ArtifactBasename, result.MemberCount)
	case restore.ModeDryRun:
		_, _ = fmt.Fprintf(os.Stdout, "restore dry-run succeeded: artifact=%s scope=%s\n",
			result.ArtifactBasename, result.Scope)
	case restore.ModeRestore:
		_, _ = fmt.Fprintf(os.Stdout, "restore succeeded: artifact=%s preserve=%s members=%d\n",
			result.ArtifactBasename, result.PreserveBasename, result.MemberCount)
		_, _ = fmt.Fprintln(os.Stdout, "restore note: web sessions require re-authentication; this restored appliance state only")
	}
	return 0
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, strings.TrimSpace(`usage:
  vyntrio-restore validate <artifact-basename>
  vyntrio-restore <artifact-basename> --dry-run
  vyntrio-restore <artifact-basename> --force

Artifact basename must name a completed backup under /var/lib/vyntrio/backups/.
Installing or restoring Vyntrio OS on hardware is a separate procedure.`))
}
