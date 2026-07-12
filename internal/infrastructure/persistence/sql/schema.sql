CREATE TABLE IF NOT EXISTS schema_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS settings (
    namespace TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL DEFAULT 'string'
        CHECK (value_type IN ('string', 'int', 'bool')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (namespace, key)
);
