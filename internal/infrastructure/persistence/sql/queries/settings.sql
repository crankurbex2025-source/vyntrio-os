-- name: GetSetting :one
SELECT namespace, key, value, value_type, updated_at
FROM settings
WHERE namespace = ? AND key = ?;

-- name: ListSettingsByNamespace :many
SELECT namespace, key, value, value_type, updated_at
FROM settings
WHERE namespace = ?
ORDER BY key;

-- name: UpsertSetting :exec
INSERT INTO settings (namespace, key, value, value_type, updated_at)
VALUES (?, ?, ?, ?, datetime('now'))
ON CONFLICT (namespace, key) DO UPDATE SET
    value = excluded.value,
    value_type = excluded.value_type,
    updated_at = datetime('now');
