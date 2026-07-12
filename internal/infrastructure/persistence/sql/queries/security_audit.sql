-- name: AppendSecurityAuditEvent :exec
INSERT INTO security_audit_events (
    id, actor_user_id, subject_user_id, event_type, result,
    ip_hash, user_agent_hash, metadata_json, occurred_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, datetime('now')
);

-- name: ListSecurityAuditEvents :many
SELECT
    id, occurred_at, actor_user_id, subject_user_id, event_type, result,
    ip_hash, user_agent_hash, metadata_json
FROM security_audit_events
WHERE (
    sqlc.narg(after_occurred_at) IS NULL
    OR occurred_at < sqlc.narg(after_occurred_at)
    OR (
        occurred_at = sqlc.narg(after_occurred_at)
        AND id < sqlc.narg(after_id)
    )
)
ORDER BY occurred_at DESC, id DESC
LIMIT sqlc.arg(row_limit);
