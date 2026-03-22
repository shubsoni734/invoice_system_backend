-- name: ListRoles :many
SELECT id, organisation_id, name, description, is_system, created_at, updated_at
FROM roles
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: GetRoleByID :one
SELECT id, organisation_id, name, description, is_system, created_at, updated_at
FROM roles
WHERE id = $1 AND organisation_id = $2;

-- name: GetRoleByName :one
SELECT id, organisation_id, name, description, is_system, created_at, updated_at
FROM roles
WHERE organisation_id = $1 AND name = $2;

-- name: CreateRole :one
INSERT INTO roles (organisation_id, name, description, is_system, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, organisation_id, name, description, is_system, created_at, updated_at;

-- name: UpdateRole :one
UPDATE roles
SET name = $3, description = $4, updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING id, organisation_id, name, description, is_system, created_at, updated_at;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1 AND organisation_id = $2 AND is_system = false;
