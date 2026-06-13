-- name: GetCustomers :many
SELECT id, organisation_id, customer_type, name, company_name, email, phone,
    gst_treatment, tax_number, business_legal_name, business_trade_name,
    place_of_supply, pan, tax_preference, exemption_reason, gst_tds,
    payment_term, document, address, contact_person_name, contact_person_email,
    contact_person_mobile, remark, is_active, created_at, updated_at
FROM customers
WHERE organisation_id = $1
AND (
    sqlc.arg('search')::TEXT = '' OR
    name ILIKE '%' || sqlc.arg('search') || '%' OR
    email ILIKE '%' || sqlc.arg('search') || '%' OR
    tax_number ILIKE '%' || sqlc.arg('search') || '%'
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
    tax_number ILIKE '%' || sqlc.arg('search') || '%'
);

-- name: GetCustomerByID :one
SELECT id, organisation_id, customer_type, name, company_name, email, phone,
    gst_treatment, tax_number, business_legal_name, business_trade_name,
    place_of_supply, pan, tax_preference, exemption_reason, gst_tds,
    payment_term, document, address, contact_person_name, contact_person_email,
    contact_person_mobile, remark, is_active, created_at, updated_at
FROM customers
WHERE id = $1 AND organisation_id = $2;

-- name: CreateCustomer :one
INSERT INTO customers (
    organisation_id, customer_type, name, company_name, email, phone,
    gst_treatment, tax_number, business_legal_name, business_trade_name,
    place_of_supply, pan, tax_preference, exemption_reason, gst_tds,
    payment_term, document, address, contact_person_name, contact_person_email,
    contact_person_mobile, remark, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23
)
RETURNING id, organisation_id, customer_type, name, company_name, email, phone,
    gst_treatment, tax_number, business_legal_name, business_trade_name,
    place_of_supply, pan, tax_preference, exemption_reason, gst_tds,
    payment_term, document, address, contact_person_name, contact_person_email,
    contact_person_mobile, remark, is_active, created_at, updated_at;

-- name: UpdateCustomer :one
UPDATE customers
SET
    customer_type        = COALESCE(sqlc.narg('customer_type'), customer_type),
    name                 = COALESCE(sqlc.narg('name'), name),
    company_name         = COALESCE(sqlc.narg('company_name'), company_name),
    email                = COALESCE(sqlc.narg('email'), email),
    phone                = COALESCE(sqlc.narg('phone'), phone),
    gst_treatment        = COALESCE(sqlc.narg('gst_treatment'), gst_treatment),
    tax_number           = COALESCE(sqlc.narg('tax_number'), tax_number),
    business_legal_name  = COALESCE(sqlc.narg('business_legal_name'), business_legal_name),
    business_trade_name  = COALESCE(sqlc.narg('business_trade_name'), business_trade_name),
    place_of_supply      = COALESCE(sqlc.narg('place_of_supply'), place_of_supply),
    pan                  = COALESCE(sqlc.narg('pan'), pan),
    tax_preference       = COALESCE(sqlc.narg('tax_preference'), tax_preference),
    exemption_reason     = COALESCE(sqlc.narg('exemption_reason'), exemption_reason),
    gst_tds              = COALESCE(sqlc.narg('gst_tds'), gst_tds),
    payment_term         = COALESCE(sqlc.narg('payment_term'), payment_term),
    document             = COALESCE(sqlc.narg('document'), document),
    address              = COALESCE(sqlc.narg('address'), address),
    contact_person_name  = COALESCE(sqlc.narg('contact_person_name'), contact_person_name),
    contact_person_email = COALESCE(sqlc.narg('contact_person_email'), contact_person_email),
    contact_person_mobile = COALESCE(sqlc.narg('contact_person_mobile'), contact_person_mobile),
    remark               = COALESCE(sqlc.narg('remark'), remark),
    is_active            = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at           = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING id, organisation_id, customer_type, name, company_name, email, phone,
    gst_treatment, tax_number, business_legal_name, business_trade_name,
    place_of_supply, pan, tax_preference, exemption_reason, gst_tds,
    payment_term, document, address, contact_person_name, contact_person_email,
    contact_person_mobile, remark, is_active, created_at, updated_at;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1 AND organisation_id = $2;
