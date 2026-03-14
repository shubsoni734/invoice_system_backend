-- +goose Up
CREATE TABLE customers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    email           TEXT,
    phone           TEXT,
    address         TEXT,
    tax_number      TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_customers_org   ON customers(organisation_id);
CREATE INDEX idx_customers_email ON customers(organisation_id, email);
CREATE INDEX idx_customers_phone ON customers(organisation_id, phone);

-- +goose Down
DROP TABLE IF EXISTS customers;
