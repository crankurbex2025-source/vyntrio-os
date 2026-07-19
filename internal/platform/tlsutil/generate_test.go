package tlsutil_test

import (
	"os"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/tlsutil"
)

func TestWriteSelfSignedFiles(t *testing.T) {
	dir := t.TempDir()
	cert, key, err := tlsutil.WriteSelfSignedFiles(dir, "127.0.0.1", "192.168.1.10")
	if err != nil {
		t.Fatalf("WriteSelfSignedFiles() error: %v", err)
	}
	for _, path := range []string{cert, key} {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			t.Fatalf("expected readable TLS file at %s", path)
		}
	}
}
