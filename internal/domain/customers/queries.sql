-- name: GetCustomers :many
SELECT *
FROM customers
WHERE organisation_id = $1
AND (
    sqlc.arg('search')::TEXT = '' OR 
    name ILIKE '%' || sqlc.arg('search') || '%' OR 
    email ILIKE '%' || sqlc.arg('search') || '%' OR 
    phone ILIKE '%' || sqlc.arg('search') || '%'
)
ORDER BY created_at DESC
LIMIT sqlc.arg('per_page') OFFSET sqlc.arg('offset');

-- name: CountCustomers :one
SELECT COUNT(*)
FROM customers
WHERE organisation_id = $1
AND (
    sqlc.arg('search')::TEXT = '' OR 
    name ILIKE '%' || sqlc.arg('search') || '%' OR 
    email ILIKE '%' || sqlc.arg('search') || '%' OR 
    phone ILIKE '%' || sqlc.arg('search') || '%'
);

-- name: GetCustomerByID :one
SELECT *
FROM customers
WHERE id = $1 AND organisation_id = $2;

-- name: CreateCustomer :one
INSERT INTO customers (
    organisation_id, name, email, phone, address, tax_number, is_active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
)
RETURNING *;

-- name: UpdateCustomer :one
UPDATE customers
SET
    name = COALESCE(sqlc.narg('name'), name),
    email = COALESCE(sqlc.narg('email'), email),
    phone = COALESCE(sqlc.narg('phone'), phone),
    address = COALESCE(sqlc.narg('address'), address),
    tax_number = COALESCE(sqlc.narg('tax_number'), tax_number),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1 AND organisation_id = $2;
