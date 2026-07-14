// Package main is the root-only local backup command for Vyntrio OS.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backup"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	runner := backup.Runner{
		RootChecker:     backup.OSRootChecker{},
		Service:         backup.SystemdController{},
		Health:          backup.LoopbackHealthProber{},
		VersionFetcher:  backup.HTTPVersionFetcher{},
		MigrationReader: backup.SQLMigrationReader{},
		Clock:           backup.RealClock(),
		Stdout:          os.Stdout,
	}

	result, err := runner.Run(ctx)
	if err != nil {
		if result.ArtifactName == "" {
			_, _ = fmt.Fprintf(os.Stderr, "backup failed\n")
		}
		return 1
	}
	_ = result
	return 0
}
