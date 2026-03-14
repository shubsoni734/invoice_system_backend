-- +goose Up
CREATE TABLE plans (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                    TEXT NOT NULL UNIQUE,
    price_monthly           NUMERIC(10,2) NOT NULL DEFAULT 0,
    price_yearly            NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_users               INT NOT NULL DEFAULT 1,
    max_customers           INT NOT NULL DEFAULT 100,
    max_invoices_per_month  INT NOT NULL DEFAULT 50,
    max_storage_mb          INT NOT NULL DEFAULT 500,
    whatsapp_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    custom_templates        BOOLEAN NOT NULL DEFAULT FALSE,
    api_access              BOOLEAN NOT NULL DEFAULT FALSE,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS plans;
