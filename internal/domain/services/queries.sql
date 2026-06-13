-- name: CountServices :one
SELECT COUNT(*)
FROM services
WHERE organisation_id = $1
AND (
    sqlc.arg('search')::TEXT = '' OR
    name     ILIKE '%' || sqlc.arg('search') || '%' OR
    hsn_code ILIKE '%' || sqlc.arg('search') || '%' OR
    sac_code ILIKE '%' || sqlc.arg('search') || '%'
);

-- name: GetServices :many
SELECT id, organisation_id, type, name, description, unit, hsn_code, sac_code,
    tax_preference, intra_state_tax_rate, inter_state_tax_rate, selling_price,
    account, is_active, created_at, updated_at
FROM services
WHERE organisation_id = $1
AND (
    sqlc.arg('search')::TEXT = '' OR
    name     ILIKE '%' || sqlc.arg('search') || '%' OR
    hsn_code ILIKE '%' || sqlc.arg('search') || '%' OR
    sac_code ILIKE '%' || sqlc.arg('search') || '%'
)
ORDER BY created_at DESC
LIMIT sqlc.arg('per_page') OFFSET sqlc.arg('offset');

-- name: GetServiceByID :one
SELECT id, organisation_id, type, name, description, unit, hsn_code, sac_code,
    tax_preference, intra_state_tax_rate, inter_state_tax_rate, selling_price,
    account, is_active, created_at, updated_at
FROM services
WHERE id = $1 AND organisation_id = $2;

-- name: CreateService :one
INSERT INTO services (
    organisation_id, type, name, description, unit, hsn_code, sac_code,
    tax_preference, intra_state_tax_rate, inter_state_tax_rate, selling_price,
    account, is_active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW()
)
RETURNING id, organisation_id, type, name, description, unit, hsn_code, sac_code,
    tax_preference, intra_state_tax_rate, inter_state_tax_rate, selling_price,
    account, is_active, created_at, updated_at;

-- name: UpdateService :one
UPDATE services
SET
    type                 = COALESCE(sqlc.narg('type'),                 type),
    name                 = COALESCE(sqlc.narg('name'),                 name),
    description          = COALESCE(sqlc.narg('description'),          description),
    unit                 = COALESCE(sqlc.narg('unit'),                 unit),
    hsn_code             = COALESCE(sqlc.narg('hsn_code'),             hsn_code),
    sac_code             = COALESCE(sqlc.narg('sac_code'),             sac_code),
    tax_preference       = COALESCE(sqlc.narg('tax_preference'),       tax_preference),
    intra_state_tax_rate = COALESCE(sqlc.narg('intra_state_tax_rate'), intra_state_tax_rate),
    inter_state_tax_rate = COALESCE(sqlc.narg('inter_state_tax_rate'), inter_state_tax_rate),
    selling_price        = COALESCE(sqlc.narg('selling_price'),        selling_price),
    account              = COALESCE(sqlc.narg('account'),              account),
    is_active            = COALESCE(sqlc.narg('is_active'),            is_active),
    updated_at           = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING id, organisation_id, type, name, description, unit, hsn_code, sac_code,
    tax_preference, intra_state_tax_rate, inter_state_tax_rate, selling_price,
    account, is_active, created_at, updated_at;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = $1 AND organisation_id = $2;
