package backup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func withRootOwnerUID(t *testing.T) func() {
	t.Helper()
	orig := destinationOwnerUID
	destinationOwnerUID = func(os.FileInfo) (uint32, bool) { return 0, true }
	return func() { destinationOwnerUID = orig }
}

func TestEnsureDestinationCreatesMissingWithMode0700(t *testing.T) {
	defer withRootOwnerUID(t)()
	path := filepath.Join(t.TempDir(), "backups")
	if err := EnsureDestination(path); err != nil {
		t.Fatalf("EnsureDestination() error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory")
	}
	if info.Mode().Perm() != destinationMode {
		t.Fatalf("mode = %o, want %o", info.Mode().Perm(), destinationMode)
	}
}

func TestEnsureDestinationAcceptsRootOwned0700Directory(t *testing.T) {
	defer withRootOwnerUID(t)()
	path := filepath.Join(t.TempDir(), "backups")
	if err := os.Mkdir(path, destinationMode); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDestination(path); err != nil {
		t.Fatalf("EnsureDestination() error: %v", err)
	}
}

func TestEnsureDestinationRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, destinationMode); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "backups")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDestination(link); !errors.Is(err, ErrDestinationUnsafe) {
		t.Fatalf("err = %v, want ErrDestinationUnsafe", err)
	}
}

func TestEnsureDestinationRejectsNonDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDestination(path); !errors.Is(err, ErrDestinationUnsafe) {
		t.Fatalf("err = %v, want ErrDestinationUnsafe", err)
	}
}

func TestEnsureDestinationRejectsMorePermissiveMode(t *testing.T) {
	defer withRootOwnerUID(t)()
	path := filepath.Join(t.TempDir(), "backups")
	if err := os.Mkdir(path, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDestination(path); !errors.Is(err, ErrDestinationUnsafe) {
		t.Fatalf("err = %v, want ErrDestinationUnsafe", err)
	}
}

func TestEnsureDestinationRejectsLessPermissiveMode(t *testing.T) {
	defer withRootOwnerUID(t)()
	path := filepath.Join(t.TempDir(), "backups")
	if err := os.Mkdir(path, 0o500); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDestination(path); !errors.Is(err, ErrDestinationUnsafe) {
		t.Fatalf("err = %v, want ErrDestinationUnsafe", err)
	}
}

func TestEnsureDestinationRejectsNonRootOwnership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backups")
	if err := os.Mkdir(path, destinationMode); err != nil {
		t.Fatal(err)
	}
	orig := destinationOwnerUID
	destinationOwnerUID = func(os.FileInfo) (uint32, bool) { return 1000, true }
	defer func() { destinationOwnerUID = orig }()
	if err := EnsureDestination(path); !errors.Is(err, ErrDestinationUnsafe) {
		t.Fatalf("err = %v, want ErrDestinationUnsafe", err)
	}
}
