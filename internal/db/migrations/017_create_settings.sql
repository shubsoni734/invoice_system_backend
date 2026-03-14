-- +goose Up
CREATE TABLE settings (
    id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id           UUID NOT NULL UNIQUE REFERENCES organisations(id) ON DELETE CASCADE,
    business_name             TEXT,
    business_email            TEXT,
    business_phone            TEXT,
    business_address          TEXT,
    logo_url                  TEXT,
    currency                  TEXT NOT NULL DEFAULT 'USD',
    date_format               TEXT NOT NULL DEFAULT 'YYYY-MM-DD',
    invoice_prefix            TEXT NOT NULL DEFAULT 'INV',
    default_due_days          INT NOT NULL DEFAULT 30,
    default_tax_rate          NUMERIC(5,2) NOT NULL DEFAULT 0,
    default_template_id       UUID,
    whatsapp_enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    whatsapp_api_key          TEXT,
    whatsapp_message_template TEXT NOT NULL DEFAULT 'Dear {{customer_name}}, please find invoice {{invoice_number}} for {{total}}. Due: {{due_date}}.',
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS settings;
