-- +goose Up
CREATE TABLE IF NOT EXISTS expenses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    category        TEXT NOT NULL,
    vendor_name     TEXT NOT NULL,
    gst_number      TEXT,
    amount          NUMERIC(12,2) NOT NULL DEFAULT 0,
    gst_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(12,2) NOT NULL DEFAULT 0,
    payment_mode    TEXT NOT NULL DEFAULT 'cash',
    payment_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    notes           TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_expenses_org_date ON expenses(organisation_id, payment_date DESC);

-- +goose Down
DROP TABLE IF EXISTS expenses;
