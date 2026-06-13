-- +goose Up

-- Remove all dependent data first, then drop and recreate customers with new schema
DROP TABLE IF EXISTS customers CASCADE;

CREATE TABLE customers (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id       UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    customer_type         VARCHAR(20)  NOT NULL DEFAULT 'business',
    name                  VARCHAR(255) NOT NULL,
    company_name          VARCHAR(255),
    email                 VARCHAR(255),
    phone                 VARCHAR(30),
    gst_treatment         VARCHAR(100),
    tax_number            VARCHAR(20),
    business_legal_name   VARCHAR(255),
    business_trade_name   VARCHAR(255),
    place_of_supply       VARCHAR(100),
    pan                   VARCHAR(10),
    tax_preference        VARCHAR(20) NOT NULL DEFAULT 'taxable',
    exemption_reason      TEXT,
    gst_tds               BOOLEAN NOT NULL DEFAULT FALSE,
    payment_term          VARCHAR(100),
    document              TEXT,
    address               TEXT,
    contact_person_name   VARCHAR(255),
    contact_person_email  VARCHAR(255),
    contact_person_mobile VARCHAR(20),
    remark                TEXT,
    is_active             BOOLEAN NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customers_org        ON customers(organisation_id);
CREATE INDEX idx_customers_email      ON customers(organisation_id, email);
CREATE INDEX idx_customers_tax_number ON customers(organisation_id, tax_number);

-- Re-add FK constraint from invoices to new customers table
ALTER TABLE invoices
    ADD CONSTRAINT invoices_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES customers(id);

-- +goose Down
DROP TABLE IF EXISTS customers CASCADE;

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

ALTER TABLE invoices
    ADD CONSTRAINT invoices_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES customers(id);
