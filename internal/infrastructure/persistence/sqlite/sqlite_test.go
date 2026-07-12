package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite"
	"github.com/crankurbex2025-source/vyntrio-os/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
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

func TestIdentityTablesExistAfterMigrate(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	for _, table := range []string{"users", "sessions", "security_audit_events"} {
		if !tableExists(t, store.DB(), table) {
			t.Fatalf("table %q missing after migrate", table)
		}
	}

	for _, idx := range []string{
		"idx_sessions_user_id",
		"idx_sessions_expires_at",
		"idx_sessions_user_active",
		"idx_security_audit_events_occurred_at",
		"idx_security_audit_events_actor_user_id",
		"idx_security_audit_events_subject_user_id",
		"idx_security_audit_events_event_type",
	} {
		if !indexExists(t, store.DB(), idx) {
			t.Fatalf("index %q missing after migrate", idx)
		}
	}
}

func TestIdentityRoleCheckConstraint(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	_, err := store.DB().Exec(`
		INSERT INTO users (id, username, password_hash, role)
		VALUES ('11111111-1111-1111-1111-111111111111', 'owner-user', 'hash', 'invalid_role')
	`)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Fatalf("expected role CHECK constraint failure, got %v", err)
	}
}

func TestIdentityStatusCheckConstraint(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	_, err := store.DB().Exec(`
		INSERT INTO users (id, username, password_hash, role, status)
		VALUES ('22222222-2222-2222-2222-222222222222', 'status-user', 'hash', 'user', 'pending')
	`)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Fatalf("expected status CHECK constraint failure, got %v", err)
	}
}

func TestUsernameCaseInsensitiveUnique(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	_, err := store.DB().Exec(`
		INSERT INTO users (id, username, password_hash, role)
		VALUES ('33333333-3333-3333-3333-333333333333', 'CaseUser', 'hash', 'user')
	`)
	if err != nil {
		t.Fatalf("first insert error: %v", err)
	}

	_, err = store.DB().Exec(`
		INSERT INTO users (id, username, password_hash, role)
		VALUES ('44444444-4444-4444-4444-444444444444', 'caseuser', 'hash', 'user')
	`)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("expected case-insensitive unique failure, got %v", err)
	}
}

func TestIdentityForeignKeyBehavior(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()
	db := store.DB()

	assertForeignKeysEnabled(t, db)

	const userID = "55555555-5555-5555-5555-555555555555"
	_, err := db.Exec(`
		INSERT INTO users (id, username, password_hash, role)
		VALUES (?, 'fk-user', 'hash', 'user')
	`, userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sessions (
			id, session_token_hash, user_id, csrf_token_hash,
			expires_at, idle_expires_at
		) VALUES (
			'66666666-6666-6666-6666-666666666666', 'session-hash', ?, 'csrf-hash',
			datetime('now', '+1 day'), datetime('now', '+1 day')
		)
	`, userID)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	_, err = db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "foreign key") {
		t.Fatalf("expected FK RESTRICT on user delete with active session, got %v", err)
	}

	if _, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	const auditID = "77777777-7777-7777-7777-777777777777"
	_, err = db.Exec(`
		INSERT INTO security_audit_events (
			id, actor_user_id, subject_user_id, event_type, result
		) VALUES (?, ?, ?, 'user.created', 'success')
	`, auditID, userID, userID)
	if err != nil {
		t.Fatalf("insert audit event: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, userID); err != nil {
		t.Fatalf("delete user with audit references: %v", err)
	}

	var actorID, subjectID sql.NullString
	err = db.QueryRow(`
		SELECT actor_user_id, subject_user_id
		FROM security_audit_events
		WHERE id = ?
	`, auditID).Scan(&actorID, &subjectID)
	if err != nil {
		t.Fatalf("select audit event: %v", err)
	}
	if actorID.Valid || subjectID.Valid {
		t.Fatalf("expected NULL actor/subject after user delete, got actor=%v subject=%v", actorID, subjectID)
	}
	if !tableExists(t, db, "security_audit_events") {
		t.Fatal("audit event row should remain after user delete")
	}
}

func TestIdentityDownMigrationRollback(t *testing.T) {
	dir := t.TempDir()
	db, migrator := openMigratedDatabase(t, dir)
	defer db.Close()

	for _, table := range []string{"users", "sessions", "security_audit_events"} {
		if !tableExists(t, db, table) {
			t.Fatalf("table %q missing before rollback", table)
		}
	}

	if err := migrator.Steps(-1); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Steps(-1) error: %v", err)
	}

	for _, table := range []string{"users", "sessions", "security_audit_events"} {
		if tableExists(t, db, table) {
			t.Fatalf("table %q should not exist after 000003 down", table)
		}
	}

	for _, idx := range []string{
		"idx_sessions_user_id",
		"idx_sessions_expires_at",
		"idx_sessions_user_active",
		"idx_security_audit_events_occurred_at",
		"idx_security_audit_events_actor_user_id",
		"idx_security_audit_events_subject_user_id",
		"idx_security_audit_events_event_type",
	} {
		if indexExists(t, db, idx) {
			t.Fatalf("index %q should not exist after 000003 down", idx)
		}
	}

	for _, table := range []string{"schema_meta", "settings"} {
		if !tableExists(t, db, table) {
			t.Fatalf("table %q should remain after 000003 down", table)
		}
	}
}

func assertForeignKeysEnabled(t *testing.T, db *sql.DB) {
	t.Helper()
	var enabled int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&enabled); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1", enabled)
	}
}

func openMigratedDatabase(t *testing.T, dir string) (*sql.DB, *migrate.Migrate) {
	t.Helper()

	path := filepath.Join(dir, "vyntrio.db")
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", path))
	if err != nil {
		t.Fatalf("sql.Open() error: %v", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		_ = db.Close()
		t.Fatalf("sqlite3.WithInstance() error: %v", err)
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		_ = db.Close()
		t.Fatalf("iofs.New() error: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		_ = db.Close()
		t.Fatalf("migrate.NewWithInstance() error: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		_ = db.Close()
		t.Fatalf("migrate Up() error: %v", err)
	}

	return db, m
}

func openTestStore(t *testing.T) *sqlite.Store {
	t.Helper()
	store, err := sqlite.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	return store
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = ?
	`, name).Scan(&count)
	if err != nil {
		t.Fatalf("tableExists(%q): %v", name, err)
	}
	return count == 1
}

func indexExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'index' AND name = ?
	`, name).Scan(&count)
	if err != nil {
		t.Fatalf("indexExists(%q): %v", name, err)
	}
	return count == 1
}
