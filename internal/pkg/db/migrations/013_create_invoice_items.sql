-- +goose Up
CREATE TABLE invoice_items (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id  UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    service_id  UUID REFERENCES services(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    quantity    NUMERIC(10,3) NOT NULL DEFAULT 1,
    unit_price  NUMERIC(12,2) NOT NULL,
    tax_rate    NUMERIC(5,2) NOT NULL DEFAULT 0,
    tax_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    line_total  NUMERIC(12,2) NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_items_invoice ON invoice_items(invoice_id);

-- +goose Down
DROP TABLE IF EXISTS invoice_items;
