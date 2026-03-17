-- name: GetServices :many
SELECT *
FROM services
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: GetServiceByID :one
SELECT *
FROM services
WHERE id = $1 AND organisation_id = $2;

-- name: CreateService :one
INSERT INTO services (
    organisation_id, name, description, unit_price, tax_rate, unit, is_active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
)
RETURNING *;

-- name: UpdateService :one
UPDATE services
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    unit_price = COALESCE(sqlc.narg('unit_price'), unit_price),
    tax_rate = COALESCE(sqlc.narg('tax_rate'), tax_rate),
    unit = COALESCE(sqlc.narg('unit'), unit),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = $1 AND organisation_id = $2;
