-- name: GetSettings :one
SELECT *
FROM settings
WHERE organisation_id = $1;

-- name: UpsertSettings :one
INSERT INTO settings (
    organisation_id, business_name, business_email, business_phone, business_address, 
    logo_url, currency, date_format, invoice_prefix, default_due_days, 
    default_tax_rate, default_template_id, whatsapp_enabled, whatsapp_api_key, 
    whatsapp_message_template, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()
)
ON CONFLICT (organisation_id)
DO UPDATE SET
    business_name = EXCLUDED.business_name,
    business_email = EXCLUDED.business_email,
    business_phone = EXCLUDED.business_phone,
    business_address = EXCLUDED.business_address,
    logo_url = EXCLUDED.logo_url,
    currency = EXCLUDED.currency,
    date_format = EXCLUDED.date_format,
    invoice_prefix = EXCLUDED.invoice_prefix,
    default_due_days = EXCLUDED.default_due_days,
    default_tax_rate = EXCLUDED.default_tax_rate,
    default_template_id = EXCLUDED.default_template_id,
    whatsapp_enabled = EXCLUDED.whatsapp_enabled,
    whatsapp_api_key = EXCLUDED.whatsapp_api_key,
    whatsapp_message_template = EXCLUDED.whatsapp_message_template,
    updated_at = NOW()
RETURNING *;
