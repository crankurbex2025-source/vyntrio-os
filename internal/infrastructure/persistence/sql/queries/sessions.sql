-- name: CreateSession :exec
INSERT INTO sessions (
    id, session_token_hash, user_id, csrf_token_hash,
    expires_at, idle_expires_at, user_agent_hash, ip_hash,
    created_at, last_seen_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetSessionByTokenHash :one
SELECT
    id, user_id, session_token_hash, csrf_token_hash,
    created_at, expires_at, idle_expires_at, last_seen_at, revoked_at,
    user_agent_hash, ip_hash
FROM sessions
WHERE session_token_hash = ?;

-- name: TouchSession :exec
UPDATE sessions
SET last_seen_at = ?
WHERE id = ?;

-- name: RevokeSessionByID :exec
UPDATE sessions
SET revoked_at = ?
WHERE id = ?;

-- name: RevokeAllSessionsForUser :exec
UPDATE sessions
SET revoked_at = ?
WHERE user_id = ? AND revoked_at IS NULL;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions
WHERE expires_at < datetime('now')
   OR idle_expires_at < datetime('now');
