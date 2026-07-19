package installpreflight

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
)

func (c Checker) checkMedia(req MediaRequest) (MediaResult, error) {
	result := MediaResult{
		EnvelopeStatus:   MediaSkipped,
		ReleaseIntegrity: ReleaseSkipped,
	}

	envelopeRoot := strings.TrimSpace(req.EnvelopeRoot)
	releaseManifest := strings.TrimSpace(req.ReleaseManifestPath)

	if envelopeRoot == "" && releaseManifest == "" {
		return result, nil
	}

	var failures []CheckFailure

	if envelopeRoot != "" {
		envelopeFailures := checkInstallEnvelope(envelopeRoot)
		if len(envelopeFailures) > 0 {
			result.EnvelopeStatus = MediaFailed
			failures = append(failures, envelopeFailures...)
		} else {
			result.EnvelopeStatus = MediaOK
		}
	}

	if releaseManifest != "" {
		baseDir := strings.TrimSpace(req.ArtifactBaseDir)
		verifyResult, err := c.Verifier.VerifyManifestFile(releaseManifest, baseDir)
		result.ReleaseIntegrity = ReleaseOK
		result.ReleaseAuth = verifyResult.Authenticity
		if err != nil {
			result.ReleaseIntegrity = ReleaseFailed
			if len(verifyResult.Failures) > 0 {
				for _, failure := range verifyResult.Failures {
					failures = append(failures, CheckFailure{
						Scope:  "release",
						Reason: failure.Reason,
						Detail: fmt.Sprintf("%s: %s", failure.Artifact, failure.Detail),
					})
				}
			} else {
				failures = append(failures, CheckFailure{
					Scope:  "release",
					Reason: releaseFailureReason(err),
					Detail: releaseManifest,
				})
			}
		}
	}

	result.Failures = failures
	if len(failures) > 0 {
		return result, fmt.Errorf("%w", ErrPreflightFailed)
	}
	return result, nil
}

func checkInstallEnvelope(envelopeRoot string) []CheckFailure {
	info, err := os.Stat(envelopeRoot)
	if err != nil || !info.IsDir() {
		return []CheckFailure{{
			Scope:  "media",
			Reason: ReasonEnvelopeMissing,
			Detail: envelopeRoot,
		}}
	}

	recordPath := filepath.Join(envelopeRoot, "ENVELOPE.txt")
	record, err := os.ReadFile(recordPath)
	if err != nil {
		return []CheckFailure{{
			Scope:  "media",
			Reason: ReasonEnvelopeRecordMissing,
			Detail: recordPath,
		}}
	}
	if !strings.Contains(string(record), "media_role: install") {
		return []CheckFailure{{
			Scope:  "media",
			Reason: ReasonEnvelopeRoleMismatch,
			Detail: "expected media_role: install",
		}}
	}

	payloadRoot := filepath.Join(envelopeRoot, "payload")
	payloadInfo, err := os.Stat(payloadRoot)
	if err != nil || !payloadInfo.IsDir() {
		return []CheckFailure{{
			Scope:  "media",
			Reason: ReasonPayloadRootMissing,
			Detail: payloadRoot,
		}}
	}

	for _, rel := range requiredPayloadFiles {
		path := filepath.Join(payloadRoot, filepath.FromSlash(rel))
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return []CheckFailure{{
				Scope:  "media",
				Reason: ReasonPayloadFileMissing,
				Detail: rel,
			}}
		}
	}

	files, err := listPayloadFiles(payloadRoot)
	if err != nil {
		return []CheckFailure{{
			Scope:  "media",
			Reason: ReasonPayloadInventoryMissing,
			Detail: err.Error(),
		}}
	}
	if len(files) != len(requiredPayloadFiles) {
		return []CheckFailure{{
			Scope:  "media",
			Reason: ReasonPayloadInventoryMissing,
			Detail: fmt.Sprintf("expected %d files, found %d", len(requiredPayloadFiles), len(files)),
		}}
	}

	allowed := make(map[string]struct{}, len(requiredPayloadFiles))
	for _, rel := range requiredPayloadFiles {
		allowed[rel] = struct{}{}
	}
	for _, rel := range files {
		if _, ok := allowed[rel]; !ok {
			return []CheckFailure{{
				Scope:  "media",
				Reason: ReasonPayloadInventoryExtra,
				Detail: rel,
			}}
		}
	}

	if failure := checkExcludedPayload(payloadRoot); failure != nil {
		return []CheckFailure{*failure}
	}
	return nil
}

func listPayloadFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

func checkExcludedPayload(payloadRoot string) *CheckFailure {
	excludedGlobs := []string{"*.db", "*.sqlite", "*.sqlite3", "*.tar", "*.tar.gz", "*.tgz"}
	for _, pattern := range excludedGlobs {
		matches, _ := filepath.Glob(filepath.Join(payloadRoot, filepath.FromSlash(pattern)))
		if len(matches) > 0 {
			return &CheckFailure{
				Scope:  "media",
				Reason: ReasonPayloadExcludedPattern,
				Detail: pattern,
			}
		}
	}

	files, err := listPayloadFiles(payloadRoot)
	if err != nil {
		return &CheckFailure{
			Scope:  "media",
			Reason: ReasonPayloadInventoryMissing,
			Detail: err.Error(),
		}
	}
	for _, rel := range files {
		base := strings.ToLower(filepath.Base(rel))
		if strings.Contains(base, "license") || (strings.Contains(base, "bootstrap") && strings.Contains(base, "token")) {
			return &CheckFailure{
				Scope:  "media",
				Reason: ReasonPayloadExcludedBasename,
				Detail: rel,
			}
		}
		if base == "vyntrio-restore" {
			return &CheckFailure{
				Scope:  "media",
				Reason: ReasonPayloadExcludedBasename,
				Detail: rel,
			}
		}
	}
	return nil
}

func releaseFailureReason(err error) string {
	switch {
	case errors.Is(err, releaseartifact.ErrVerifyFailed):
		return releaseartifact.ReasonSHA256Mismatch
	case errors.Is(err, releaseartifact.ErrMalformedManifest):
		return releaseartifact.ReasonMalformedManifest
	default:
		return releaseartifact.ReasonMalformedManifest
	}
}
