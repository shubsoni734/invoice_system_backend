-- name: GetPaymentsByInvoice :many
SELECT *
FROM payments
WHERE invoice_id = $1 AND organisation_id = $2
ORDER BY created_at DESC;

-- name: GetPaymentsByOrg :many
SELECT *
FROM payments
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: CreatePayment :one
INSERT INTO payments (
    organisation_id, invoice_id, amount, method, reference, notes, payment_date, recorded_by, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
RETURNING *;

-- name: DeletePayment :exec
DELETE FROM payments
WHERE id = $1 AND organisation_id = $2;
