-- name: OrgSlugExists :one
SELECT EXISTS(SELECT 1 FROM organisations WHERE slug = $1)::boolean;

-- name: OrgExists :one
SELECT EXISTS(SELECT 1 FROM organisations WHERE id = $1)::boolean;

-- name: CreateOrganisation :one
INSERT INTO organisations (name, slug, email, phone, address, status, created_by_super_admin_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, 'active', $6, NOW(), NOW())
RETURNING id, name, slug, email, created_at;

-- name: CountOrganisations :one
SELECT COUNT(*)::int FROM organisations;

-- name: ListOrganisations :many
SELECT id, name, slug, email, phone, status, created_at
FROM organisations
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetOrganisationByID :one
SELECT id, name, slug, email, phone, address, status, created_at, updated_at
FROM organisations
WHERE id = $1;

-- name: GetActiveSubscription :one
SELECT p.name AS plan_name, os.status, os.current_period_end
FROM organisation_subscriptions os
JOIN plans p ON p.id = os.plan_id
WHERE os.organisation_id = $1 AND os.status = 'active'
ORDER BY os.created_at DESC
LIMIT 1;

-- name: GetPlanByID :one
SELECT id FROM plans WHERE id = $1 AND is_active = true;

-- name: GetUnlimitedPlan :one
SELECT id FROM plans WHERE name = 'Unlimited' LIMIT 1;

-- name: CreateUnlimitedPlan :one
INSERT INTO plans (name, price_monthly, price_yearly, max_users, max_customers,
    max_invoices_per_month, max_storage_mb, whatsapp_enabled, custom_templates,
    api_access, is_active, created_at, updated_at)
VALUES ('Unlimited', 0, 0, 999999, 999999, 999999, 999999, true, true, true, true, NOW(), NOW())
RETURNING id;

-- name: CancelActiveSubscriptions :exec
UPDATE organisation_subscriptions
SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
WHERE organisation_id = $1 AND status = 'active';

-- name: CreateSubscription :one
INSERT INTO organisation_subscriptions
    (organisation_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at)
VALUES ($1, $2, 'active', NOW(), NOW() + INTERVAL '100 years', NOW(), NOW())
RETURNING id;
