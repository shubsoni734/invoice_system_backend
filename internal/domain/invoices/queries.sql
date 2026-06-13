-- name: CountInvoices :one
SELECT COUNT(*)
FROM invoices
WHERE organisation_id = $1
  AND ($2::TEXT = '' OR status = $2)
  AND ($3::UUID IS NULL OR customer_id = $3::UUID)
  AND ($4::DATE IS NULL OR issued_date >= $4::DATE)
  AND ($5::DATE IS NULL OR issued_date <= $5::DATE);

-- name: GetInvoices :many
SELECT i.id, i.organisation_id, i.customer_id, i.session_id, i.invoice_number, i.status,
    i.issued_date, i.due_date, i.subtotal, i.tax_amount, i.cgst_amount, i.sgst_amount, i.igst_amount,
    i.discount_amount, i.total, i.grand_total, i.is_same_state, i.currency, i.notes, i.terms,
    i.template_id, i.created_by, i.sent_at, i.paid_at, i.created_at, i.updated_at,
    c.name AS customer_name, c.tax_number AS customer_gstin
FROM invoices i
LEFT JOIN customers c ON i.customer_id = c.id
WHERE i.organisation_id = $1
  AND ($2::TEXT = '' OR i.status = $2)
  AND ($3::UUID IS NULL OR i.customer_id = $3::UUID)
  AND ($4::DATE IS NULL OR i.issued_date >= $4::DATE)
  AND ($5::DATE IS NULL OR i.issued_date <= $5::DATE)
ORDER BY i.created_at DESC
LIMIT $7 OFFSET $6;

-- name: GetInvoiceByID :one
SELECT id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at
FROM invoices
WHERE id = $1 AND organisation_id = $2;

-- name: GetInvoiceForValidation :one
SELECT id, status, sent_at, paid_at
FROM invoices
WHERE id = $1 AND organisation_id = $2;

-- name: IncrementSessionSequence :one
UPDATE invoice_sessions
SET current_sequence = current_sequence + 1
WHERE id = $1 AND organisation_id = $2
RETURNING current_sequence, prefix, year;

-- name: CreateInvoice :one
INSERT INTO invoices (
    organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW(), NOW()
)
RETURNING id;

-- name: CreateInvoiceItem :exec
INSERT INTO invoice_items (
    invoice_id, service_id, description, hsn_sac_code, quantity, unit_price,
    tax_rate, tax_amount, line_total, sort_order, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW()
);

-- name: GetInvoiceItems :many
SELECT id, invoice_id, service_id, description, hsn_sac_code, quantity, unit_price,
    tax_rate, tax_amount, line_total, sort_order, created_at
FROM invoice_items
WHERE invoice_id = $1
ORDER BY sort_order ASC;

-- name: CancelInvoice :one
UPDATE invoices
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 AND organisation_id = $2 AND status NOT IN ('paid')
RETURNING id, status;

-- name: DeleteInvoice :exec
DELETE FROM invoices
WHERE id = $1 AND organisation_id = $2;

-- name: SetInvoiceSent :one
UPDATE invoices
SET status = 'sent', sent_at = NOW(), updated_at = NOW()
WHERE id = $1 AND organisation_id = $2 AND status = 'draft'
RETURNING id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at;

-- name: SetInvoicePaid :one
UPDATE invoices
SET status = 'paid', paid_at = NOW(), updated_at = NOW()
WHERE id = $1 AND organisation_id = $2 AND status NOT IN ('cancelled')
RETURNING id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at;

-- name: UpdateInvoiceFields :one
UPDATE invoices
SET notes = $3,
    terms = $4,
    due_date = CASE WHEN $5::DATE IS NULL THEN due_date ELSE $5::DATE END,
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at;
