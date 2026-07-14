package backup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// SourceMember maps an archive member name to an absolute host source path.
type SourceMember struct {
	ArchiveName string
	SourcePath  string
}

type sourceSpec struct {
	archiveName string
	fileName    string
	required    bool
}

var sourceSpecs = []sourceSpec{
	{archiveName: StateDBMember, fileName: MainDBName, required: true},
	{archiveName: StateJournalMem, fileName: JournalSidecar, required: false},
	{archiveName: StateWALMember, fileName: WALSidecar, required: false},
	{archiveName: StateSHMMember, fileName: SHMSidecar, required: false},
	{archiveName: ConfigMember, fileName: "", required: true},
}

// SelectSourceMembers validates and returns the approved source members.
func SelectSourceMembers(stateRoot, configPath string) ([]SourceMember, error) {
	members := make([]SourceMember, 0, len(sourceSpecs))
	for _, spec := range sourceSpecs {
		sourcePath := filepath.Join(stateRoot, spec.fileName)
		if spec.archiveName == ConfigMember {
			sourcePath = configPath
		}
		info, err := validateRegularSource(sourcePath, spec.required)
		if err != nil {
			return nil, err
		}
		if info == nil {
			continue
		}
		members = append(members, SourceMember{
			ArchiveName: spec.archiveName,
			SourcePath:  sourcePath,
		})
	}
	return members, nil
}

func validateRegularSource(path string, required bool) (fs.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if required {
				return nil, fmt.Errorf("%w: required source missing", ErrSourceInvalid)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("%w: source inaccessible", ErrSourceInvalid)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%w: symlink source rejected", ErrSourceInvalid)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: non-regular source rejected", ErrSourceInvalid)
	}
	return info, nil
}

// RevalidateSourceMember re-checks one source immediately before read.
func RevalidateSourceMember(member SourceMember) error {
	required := member.ArchiveName == StateDBMember || member.ArchiveName == ConfigMember
	_, err := validateRegularSource(member.SourcePath, required)
	return err
}
