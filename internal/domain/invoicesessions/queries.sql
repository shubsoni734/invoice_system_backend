-- name: GetInvoiceSessions :many
SELECT *
FROM invoice_sessions
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: GetInvoiceSessionByID :one
SELECT *
FROM invoice_sessions
WHERE id = $1 AND organisation_id = $2;

-- name: CreateInvoiceSession :one
INSERT INTO invoice_sessions (
    organisation_id, year, prefix, current_sequence, created_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
RETURNING *;

-- name: UpdateInvoiceSessionSequence :one
UPDATE invoice_sessions
SET current_sequence = current_sequence + 1
WHERE id = $1 AND organisation_id = $2
RETURNING *;
