package installwrite

// BuildCopyPlan resolves the allowlisted payload copy plan from install media sources.
func BuildCopyPlan(envelopeRoot, manifestPath, artifactBaseDir string) ([]CopyEntry, string, error) {
	return buildCopyPlan(envelopeRoot, manifestPath, artifactBaseDir)
}

// CreateStateDirectories creates supported state directories under targetRoot.
func CreateStateDirectories(targetRoot string) error {
	return createStateDirectories(targetRoot)
}

// CopyPayloadFile copies one allowlisted payload entry into targetRoot.
func CopyPayloadFile(entry CopyEntry, targetRoot string) error {
	return copyPayload(entry, targetRoot)
}

// PostVerifyTarget checks copied payloads under targetRoot.
func PostVerifyTarget(targetRoot string, plan []CopyEntry) error {
	return postVerify(targetRoot, plan)
}

// WriteInstallRecordFile writes INSTALL_RECORD.txt on the target tree.
func WriteInstallRecordFile(targetRoot, diskID, releaseVersion string, plan []CopyEntry, hostMutated bool) error {
	return writeInstallRecord(targetRoot, diskID, releaseVersion, plan, hostMutated)
}
