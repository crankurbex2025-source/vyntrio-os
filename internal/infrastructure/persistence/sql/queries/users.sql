-- name: CreateUser :exec
INSERT INTO users (
    id, username, display_name, password_hash, role, status, must_change_password,
    created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
);

-- name: GetUserByID :one
SELECT
    id, username, display_name, role, status, must_change_password,
    created_at, updated_at, last_login_at
FROM users
WHERE id = ?;

-- name: GetUserByUsername :one
SELECT
    id, username, display_name, password_hash, role, status, must_change_password,
    created_at, updated_at, last_login_at
FROM users
WHERE username = ?;

-- name: UpdateUserPasswordHash :exec
UPDATE users
SET password_hash = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: UpdateUserLastLoginAt :exec
UPDATE users
SET last_login_at = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: SetUserStatus :exec
UPDATE users
SET status = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: ListUsers :many
SELECT
    id, username, display_name, role, status, must_change_password,
    created_at, updated_at, last_login_at
FROM users
WHERE (
    sqlc.arg(after_username) = ''
    OR username > sqlc.arg(after_username)
)
ORDER BY username COLLATE NOCASE ASC
LIMIT sqlc.arg(row_limit);

-- name: CountUsers :one
SELECT COUNT(*) AS count
FROM users;
