-- name: GetPaymentsByInvoice :many
SELECT id, organisation_id, invoice_id, amount, method, reference, notes, payment_date, recorded_by, created_at
FROM payments
WHERE invoice_id = $1 AND organisation_id = $2
ORDER BY created_at DESC;

-- name: GetPaymentByID :one
SELECT id, organisation_id, invoice_id, amount, method, reference, notes, payment_date, recorded_by, created_at
FROM payments
WHERE id = $1 AND organisation_id = $2;

-- name: CreatePayment :one
INSERT INTO payments (
    organisation_id, invoice_id, amount, method, reference, notes, payment_date, recorded_by, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
RETURNING id, organisation_id, invoice_id, amount, method, reference, notes, payment_date, recorded_by, created_at;

-- name: DeletePayment :exec
DELETE FROM payments
WHERE id = $1 AND organisation_id = $2;

-- name: GetInvoicePaymentSummary :one
SELECT i.grand_total, i.status, COALESCE(SUM(p.amount), 0) AS total_paid
FROM invoices i
LEFT JOIN payments p ON p.invoice_id = i.id
WHERE i.id = $1 AND i.organisation_id = $2
GROUP BY i.grand_total, i.status;

-- name: MarkInvoicePaid :exec
UPDATE invoices
SET status = 'paid', paid_at = NOW(), updated_at = NOW()
WHERE id = $1 AND organisation_id = $2;

-- name: ReverseInvoicePaidStatus :exec
UPDATE invoices
SET status = 'sent', paid_at = NULL, updated_at = NOW()
WHERE id = $1 AND organisation_id = $2 AND status = 'paid';

-- name: GetAllPaymentsWithDetails :many
SELECT
    p.id,
    p.organisation_id,
    p.invoice_id,
    p.amount,
    p.method,
    p.reference,
    p.notes,
    p.payment_date,
    p.recorded_by,
    p.created_at,
    i.invoice_number,
    c.name AS customer_name
FROM payments p
JOIN invoices i ON i.id = p.invoice_id
JOIN customers c ON c.id = i.customer_id
WHERE p.organisation_id = $1
ORDER BY p.created_at DESC;
