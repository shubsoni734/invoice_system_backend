-- +goose Up
CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id),
    invoice_id      UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount          NUMERIC(12,2) NOT NULL,
    method          TEXT NOT NULL DEFAULT 'cash',
    reference       TEXT,
    notes           TEXT,
    payment_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    recorded_by     UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_org     ON payments(organisation_id);

-- +goose Down
DROP TABLE IF EXISTS payments;
