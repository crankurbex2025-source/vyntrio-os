package installwrite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveTargetRoot(sandboxRoot, diskID, override string) (string, error) {
	sandboxAbs, err := filepath.Abs(strings.TrimSpace(sandboxRoot))
	if err != nil {
		return "", err
	}
	if sandboxAbs == "" {
		return "", fmt.Errorf("%w: empty sandbox root", ErrUnsafeTargetRoot)
	}

	target := filepath.Join(sandboxAbs, strings.TrimSpace(diskID))
	if strings.TrimSpace(override) != "" {
		target = override
	}

	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if err := assertUnderRoot(targetAbs, sandboxAbs); err != nil {
		return "", err
	}
	return targetAbs, nil
}

func assertUnderRoot(path, root string) error {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	if path == root {
		return nil
	}
	sep := string(os.PathSeparator)
	if !strings.HasPrefix(path, root+sep) {
		return fmt.Errorf("%w: %s", ErrUnsafeTargetRoot, path)
	}
	return nil
}
