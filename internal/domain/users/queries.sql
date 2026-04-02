-- name: ListUsersByOrg :many
SELECT u.id, u.organisation_id, u.email, u.name, u.role, u.is_active, u.role_id, u.created_at,
       COALESCE(r.name, u.role) as role_display_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.organisation_id = $1
ORDER BY u.created_at DESC;

-- name: CreateOrgUser :one
INSERT INTO users (organisation_id, email, password_hash, name, role, role_id, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
RETURNING id, organisation_id, email, name, role, role_id, is_active, created_at;

-- name: GetOrgUserCount :one
SELECT COUNT(*)::bigint FROM users WHERE organisation_id = $1;

-- name: GetRoleByName :one
SELECT id, organisation_id, name, description, is_system, created_at, updated_at
FROM roles
WHERE organisation_id = $1 AND name = $2;

-- name: GetRoleByID :one
SELECT id, organisation_id, name, description, is_system, created_at, updated_at
FROM roles
WHERE id = $1 AND organisation_id = $2;

-- name: SetUserStatus :one
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE id = $1 AND organisation_id = $3
RETURNING id, organisation_id, email, name, role, role_id, is_active, updated_at;

-- name: UserEmailExistsInOrg :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND organisation_id = $2)::boolean;

-- name: CreateDefaultAdminRole :one
INSERT INTO roles (organisation_id, name, description, is_system, created_at, updated_at)
VALUES ($1, 'Admin', 'Administrator with full access', true, NOW(), NOW())
RETURNING id, organisation_id, name, description, is_system, created_at, updated_at;

-- name: UpdateOrgUser :one
UPDATE users
SET name = $1, role = $2, role_id = $3, email = $4, updated_at = NOW()
WHERE id = $5 AND organisation_id = $6
RETURNING id, organisation_id, email, name, role, role_id, is_active, updated_at;
