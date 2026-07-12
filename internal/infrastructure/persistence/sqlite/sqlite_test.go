package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
)

func TestOpenMigratePingClose(t *testing.T) {
	dir := t.TempDir()

	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}

	dbPath := filepath.Join(dir, "vyntrio.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database file missing: %v", err)
	}
}

func TestOpenIdempotentMigrations(t *testing.T) {
	dir := t.TempDir()

	store1, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("first Open() error: %v", err)
	}
	_ = store1.Close()

	store2, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("second Open() error: %v", err)
	}
	_ = store2.Close()
}
