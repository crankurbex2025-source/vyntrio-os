package installpreflight

import "path/filepath"

// RequiredPayloadRelativePaths returns install-media payload paths relative to payload/.
func RequiredPayloadRelativePaths() []string {
	out := make([]string, len(requiredPayloadFiles))
	copy(out, requiredPayloadFiles)
	return out
}

// PayloadRoot returns the payload subtree for an install-media envelope.
func PayloadRoot(envelopeRoot string) string {
	if envelopeRoot == "" {
		return ""
	}
	return filepath.Join(envelopeRoot, "payload")
}
