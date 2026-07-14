package backup

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLMigrationReader reads golang-migrate schema metadata from a database file.
type SQLMigrationReader struct{}

func (SQLMigrationReader) Read(ctx context.Context, dbPath string) (ReleaseMetadata, error) {
	if dbPath == "" {
		return ReleaseMetadata{}, fmt.Errorf("database path missing")
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return ReleaseMetadata{}, err
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var version int64
	var dirty bool
	err = db.QueryRowContext(ctx, `SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
	if err != nil {
		return ReleaseMetadata{}, err
	}
	return ReleaseMetadata{
		MigrationVersion:    fmt.Sprintf("%d", version),
		MigrationDirty:      dirty,
		MigrationDirtyKnown: true,
	}, nil
}
