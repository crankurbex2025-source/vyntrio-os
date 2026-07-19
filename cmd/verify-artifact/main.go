// Package main verifies local Vyntrio release/install artifacts against a manifest.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/installpolicy"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		usage()
		return 0
	}

	fs := flag.NewFlagSet("verify-artifact", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	baseDir := fs.String("base-dir", "", "directory containing artifact files (default: manifest directory)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	remaining := fs.Args()
	if len(remaining) != 1 {
		usage()
		return 2
	}

	result, err := releaseartifact.NewVerifier().VerifyManifestFile(remaining[0], *baseDir)
	if err != nil {
		if errors.Is(err, releaseartifact.ErrMalformedManifest) {
			_, _ = fmt.Fprintf(os.Stderr, "verify-artifact failed: reason=malformed_manifest detail=%v\n", err)
			return 1
		}
		if errors.Is(err, releaseartifact.ErrVerifyFailed) {
			for _, failure := range result.Failures {
				_, _ = fmt.Fprintf(os.Stderr, "verify-artifact failed: artifact=%s reason=%s detail=%s\n",
					failure.Artifact, failure.Reason, failure.Detail)
			}
			_, _ = fmt.Fprintf(os.Stderr, "verify-artifact failed: integrity=%s authenticity=%s\n",
				result.Integrity, result.Authenticity)
			return 1
		}
		_, _ = fmt.Fprintf(os.Stderr, "verify-artifact failed: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(os.Stdout,
		"verify-artifact succeeded: format=%s release=%s artifacts=%d integrity=%s authenticity=%s\n",
		result.FormatVersion,
		result.ReleaseVersion,
		result.ArtifactCount,
		result.Integrity,
		result.Authenticity,
	)
	if result.Authenticity == releaseartifact.AuthenticityUnsupported {
		_, _ = fmt.Fprintln(os.Stderr,
			"verify-artifact note: manifest lists a signature but Ed25519 verification is not implemented; integrity verified only")
	}
	if result.Authenticity == releaseartifact.AuthenticityNotSigned {
		_, _ = fmt.Fprintln(os.Stderr,
			"verify-artifact note: no signature in manifest; integrity verified only, authenticity not established")
	}
	return 0
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, installpolicy.VerifyArtifactUsageText())
}
