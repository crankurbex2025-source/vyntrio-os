CREATE TABLE IF NOT EXISTS settings (
    namespace TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL DEFAULT 'string'
        CHECK (value_type IN ('string', 'int', 'bool')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (namespace, key)
);

INSERT INTO settings (namespace, key, value, value_type)
VALUES ('system', 'timezone', 'UTC', 'string');

INSERT INTO settings (namespace, key, value, value_type)
VALUES ('system', 'hostname', 'vyntrio', 'string');
