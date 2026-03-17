-- name: ListUsersByOrg :many
SELECT id, organisation_id, email, name, role, is_active, created_at
FROM users
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: CreateOrgUser :one
INSERT INTO users (organisation_id, email, password_hash, name, role, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
RETURNING id, organisation_id, email, name, role, is_active, created_at;

-- name: SetUserStatus :one
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, organisation_id, email, name, role, is_active, updated_at;

-- name: UserEmailExistsInOrg :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND organisation_id = $2)::boolean;
