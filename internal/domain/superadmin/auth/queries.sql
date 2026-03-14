-- name: GetSuperAdminByEmail :one
SELECT id, email, password_hash, role, is_active, failed_attempts, locked_until, last_login_at, created_at, updated_at
FROM super_admins
WHERE email = $1;

-- name: GetSuperAdminByID :one
SELECT id, email, password_hash, role, is_active, failed_attempts, locked_until, last_login_at, created_at, updated_at
FROM super_admins
WHERE id = $1;

-- name: SuperAdminEmailExists :one
SELECT EXISTS(SELECT 1 FROM super_admins WHERE email = $1)::boolean;

-- name: CreateSuperAdmin :one
INSERT INTO super_admins (email, password_hash, role, is_active, created_at, updated_at)
VALUES ($1, $2, $3, true, NOW(), NOW())
RETURNING id, email, role, created_at;

-- name: IncrementFailedAttempts :exec
UPDATE super_admins
SET failed_attempts = failed_attempts + 1,
    locked_until = CASE WHEN failed_attempts >= 4 THEN NOW() + INTERVAL '15 minutes' ELSE NULL END
WHERE id = $1;

-- name: ResetFailedAttempts :exec
UPDATE super_admins
SET failed_attempts = 0, locked_until = NULL, last_login_at = NOW()
WHERE id = $1;

-- name: CreateSuperRefreshToken :exec
INSERT INTO super_refresh_tokens (super_admin_id, token_hash, expires_at, ip_address, user_agent, created_at)
VALUES ($1, $2, $3, $4, $5, NOW());

-- name: RevokeAllSuperRefreshTokens :exec
UPDATE super_refresh_tokens
SET revoked_at = NOW()
WHERE super_admin_id = $1 AND revoked_at IS NULL;
