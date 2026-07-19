package releaseartifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Failure records one artifact verification failure with a stable reason code.
type Failure struct {
	Artifact string
	Reason   string
	Detail   string
}

// Result is the outcome of local manifest + artifact verification.
type Result struct {
	ManifestPath   string
	BaseDir        string
	FormatVersion  string
	ReleaseVersion string
	ArtifactCount  int
	Integrity      string
	Authenticity   string
	Failures       []Failure
}

// Verifier checks manifest structure and on-disk artifact integrity.
type Verifier struct {
	HashFile func(path string) (string, error)
}

// NewVerifier creates a verifier with the default SHA-256 hasher.
func NewVerifier() Verifier {
	return Verifier{HashFile: fileSHA256}
}

// VerifyManifestFile loads a manifest and verifies all listed artifacts.
func (v Verifier) VerifyManifestFile(manifestPath, baseDir string) (Result, error) {
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return Result{}, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := DecodeManifest(manifestData)
	if err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(baseDir) == "" {
		baseDir = filepath.Dir(manifestPath)
	}
	return v.VerifyDecoded(manifest, manifestPath, baseDir)
}

// VerifyDecoded verifies a parsed manifest against files under baseDir.
func (v Verifier) VerifyDecoded(manifest Manifest, manifestPath, baseDir string) (Result, error) {
	hashFile := v.HashFile
	if hashFile == nil {
		hashFile = fileSHA256
	}

	result := Result{
		ManifestPath:   manifestPath,
		BaseDir:        baseDir,
		FormatVersion:  manifest.FormatVersion,
		ReleaseVersion: manifest.Release.Version,
		ArtifactCount:  len(manifest.Artifacts),
		Integrity:      IntegrityOK,
		Authenticity:   authenticityStatus(manifest.Signature),
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return Result{}, fmt.Errorf("resolve base dir: %w", err)
	}

	for _, artifact := range manifest.Artifacts {
		absPath, err := secureJoin(absBase, artifact.RelativePath)
		if err != nil {
			result.Integrity = IntegrityFailed
			result.Failures = append(result.Failures, Failure{
				Artifact: artifact.Name,
				Reason:   ReasonInvalidRelativePath,
				Detail:   err.Error(),
			})
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil {
			result.Integrity = IntegrityFailed
			result.Failures = append(result.Failures, Failure{
				Artifact: artifact.Name,
				Reason:   ReasonMissingFile,
				Detail:   artifact.RelativePath,
			})
			continue
		}
		if info.IsDir() {
			result.Integrity = IntegrityFailed
			result.Failures = append(result.Failures, Failure{
				Artifact: artifact.Name,
				Reason:   ReasonMissingFile,
				Detail:   "path is a directory",
			})
			continue
		}
		if uint64(info.Size()) != artifact.SizeBytes {
			result.Integrity = IntegrityFailed
			result.Failures = append(result.Failures, Failure{
				Artifact: artifact.Name,
				Reason:   ReasonSizeMismatch,
				Detail:   fmt.Sprintf("expected=%d actual=%d", artifact.SizeBytes, info.Size()),
			})
			continue
		}

		sum, err := hashFile(absPath)
		if err != nil {
			return Result{}, fmt.Errorf("hash artifact %s: %w", artifact.Name, err)
		}
		if sum != artifact.SHA256 {
			result.Integrity = IntegrityFailed
			result.Failures = append(result.Failures, Failure{
				Artifact: artifact.Name,
				Reason:   ReasonSHA256Mismatch,
				Detail:   artifact.RelativePath,
			})
		}
	}

	if result.Integrity == IntegrityFailed {
		return result, fmt.Errorf("%w", ErrVerifyFailed)
	}
	return result, nil
}

func authenticityStatus(signature *SignatureMetadata) string {
	if signature == nil {
		return AuthenticityNotSigned
	}
	return AuthenticityUnsupported
}

func secureJoin(baseDir, relativePath string) (string, error) {
	cleanRel := filepath.Clean(relativePath)
	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("relative_path escapes base directory")
	}
	joined := filepath.Join(baseDir, cleanRel)
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	if absJoined != baseDir && !strings.HasPrefix(absJoined, baseDir+string(filepath.Separator)) {
		return "", fmt.Errorf("relative_path escapes base directory")
	}
	return absJoined, nil
}
