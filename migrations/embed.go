// Package migrations embeds SQL migration files for golang-migrate.
package migrations

import "embed"

// FS contains versioned SQL migrations (see docs/ADR/0003-sqlite-migrations.md).
//
//go:embed *.sql
var FS embed.FS
