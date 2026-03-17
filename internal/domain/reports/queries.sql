-- name: GetDailyReport :one
SELECT
    COUNT(i.id)::int                          AS total_invoices,
    COALESCE(SUM(i.total), 0)::numeric        AS total_revenue,
    COALESCE(SUM(i.tax_amount), 0)::numeric   AS total_tax,
    COUNT(CASE WHEN i.status = 'paid' THEN 1 END)::int     AS paid_count,
    COUNT(CASE WHEN i.status = 'pending' THEN 1 END)::int  AS pending_count,
    COUNT(CASE WHEN i.status = 'cancelled' THEN 1 END)::int AS cancelled_count
FROM invoices i
WHERE i.organisation_id = $1
  AND i.issued_date = $2::date;

-- name: GetMonthlyReport :one
SELECT
    COUNT(i.id)::int                          AS total_invoices,
    COALESCE(SUM(i.total), 0)::numeric        AS total_revenue,
    COALESCE(SUM(i.tax_amount), 0)::numeric   AS total_tax,
    COUNT(CASE WHEN i.status = 'paid' THEN 1 END)::int     AS paid_count,
    COUNT(CASE WHEN i.status = 'pending' THEN 1 END)::int  AS pending_count,
    COUNT(CASE WHEN i.status = 'cancelled' THEN 1 END)::int AS cancelled_count
FROM invoices i
WHERE i.organisation_id = $1
  AND i.issued_date >= $2::date
  AND i.issued_date <= $3::date;

-- name: GetCustomerInvoiceHistory :many
SELECT
    i.id, i.invoice_number, i.status, i.issued_date, i.due_date,
    i.subtotal, i.tax_amount, i.total, i.currency, i.created_at
FROM invoices i
WHERE i.organisation_id = $1
  AND i.customer_id = $2
ORDER BY i.created_at DESC;

-- name: GetRevenueSummary :many
SELECT
    DATE_TRUNC('month', i.issued_date)::date AS month,
    COUNT(i.id)::int                         AS total_invoices,
    COALESCE(SUM(i.total), 0)::numeric       AS total_revenue
FROM invoices i
WHERE i.organisation_id = $1
  AND i.issued_date >= NOW() - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', i.issued_date)
ORDER BY month ASC;
