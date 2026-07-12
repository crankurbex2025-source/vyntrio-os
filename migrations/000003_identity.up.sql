CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL COLLATE NOCASE UNIQUE,
    display_name TEXT,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'administrator', 'operator', 'user', 'read_only')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    must_change_password INTEGER NOT NULL DEFAULT 0 CHECK (must_change_password IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_login_at TEXT
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    session_token_hash TEXT NOT NULL UNIQUE,
    user_id TEXT NOT NULL,
    csrf_token_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL,
    idle_expires_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL DEFAULT (datetime('now')),
    revoked_at TEXT,
    user_agent_hash TEXT,
    ip_hash TEXT,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions (user_id) WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS security_audit_events (
    id TEXT PRIMARY KEY,
    occurred_at TEXT NOT NULL DEFAULT (datetime('now')),
    actor_user_id TEXT,
    subject_user_id TEXT,
    event_type TEXT NOT NULL,
    result TEXT NOT NULL CHECK (result IN ('success', 'failure', 'denied')),
    ip_hash TEXT,
    user_agent_hash TEXT,
    metadata_json TEXT,
    FOREIGN KEY (actor_user_id) REFERENCES users (id) ON DELETE SET NULL,
    FOREIGN KEY (subject_user_id) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_security_audit_events_occurred_at ON security_audit_events (occurred_at);
CREATE INDEX IF NOT EXISTS idx_security_audit_events_actor_user_id ON security_audit_events (actor_user_id);
CREATE INDEX IF NOT EXISTS idx_security_audit_events_subject_user_id ON security_audit_events (subject_user_id);
CREATE INDEX IF NOT EXISTS idx_security_audit_events_event_type ON security_audit_events (event_type);
