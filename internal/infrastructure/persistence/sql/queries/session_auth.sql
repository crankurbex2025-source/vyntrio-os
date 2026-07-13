-- name: GetSessionAuthByTokenHash :one
SELECT
    s.id AS session_id,
    s.user_id,
    s.expires_at,
    s.idle_expires_at,
    s.revoked_at,
    s.csrf_token_hash,
    u.status AS user_status,
    u.role AS user_role
FROM sessions s
INNER JOIN users u ON u.id = s.user_id
WHERE s.session_token_hash = ?;
