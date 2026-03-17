-- name: GetInvoices :many
SELECT *
FROM invoices
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: GetInvoiceByID :one
SELECT *
FROM invoices
WHERE id = $1 AND organisation_id = $2;

-- name: CreateInvoice :one
INSERT INTO invoices (
    organisation_id, customer_id, session_id, invoice_number, status, issued_date, due_date,
    subtotal, tax_amount, discount_amount, total, currency, notes, terms, template_id, created_by, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
)
RETURNING *;

-- name: CancelInvoice :one
UPDATE invoices
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: CreateInvoiceItem :one
INSERT INTO invoice_items (
    invoice_id, service_id, description, quantity, unit_price, tax_rate, tax_amount, line_total, sort_order, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
)
RETURNING *;

-- name: GetInvoiceItems :many
SELECT *
FROM invoice_items
WHERE invoice_id = $1
ORDER BY sort_order ASC;
