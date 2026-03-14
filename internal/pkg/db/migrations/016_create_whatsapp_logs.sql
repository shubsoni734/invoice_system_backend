-- +goose Up
CREATE TABLE whatsapp_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id),
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    recipient_phone TEXT NOT NULL,
    message         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wa_logs_invoice ON whatsapp_logs(invoice_id);

-- +goose Down
DROP TABLE IF EXISTS whatsapp_logs;
