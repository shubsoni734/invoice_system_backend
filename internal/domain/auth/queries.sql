-- name: GetUserByEmail :one
SELECT id, organisation_id, email, password_hash, name, role, role_id, is_active, failed_attempts, locked_until, last_login_at, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, organisation_id, email, password_hash, name, role, role_id, is_active, failed_attempts, locked_until, last_login_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: IncrementFailedAttempts :exec
UPDATE users
SET failed_attempts = failed_attempts + 1,
    locked_until = CASE WHEN failed_attempts >= 4 THEN NOW() + INTERVAL '15 minutes' ELSE NULL END,
    updated_at = NOW()
WHERE id = $1;

-- name: ResetFailedAttempts :exec
UPDATE users
SET failed_attempts = 0, locked_until = NULL, last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip_address, user_agent, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING id, user_id, token_hash, expires_at, revoked_at, ip_address, user_agent, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token_hash, expires_at, revoked_at, ip_address, user_agent, created_at
FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: CreatePasswordResetToken :one
INSERT INTO password_resets (user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING id, user_id, token_hash, expires_at, created_at;

-- name: GetPasswordResetToken :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM password_resets
WHERE token_hash = $1 AND expires_at > NOW();

-- name: DeletePasswordResetToken :exec
DELETE FROM password_resets
WHERE id = $1;

-- name: DeleteUserPasswordResets :exec
DELETE FROM password_resets
WHERE user_id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;
