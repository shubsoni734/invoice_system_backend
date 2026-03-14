-- +goose Up
CREATE TABLE invoice_sessions (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id  UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    year             INT NOT NULL,
    prefix           TEXT NOT NULL DEFAULT 'INV',
    current_sequence INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organisation_id, year, prefix)
);

-- +goose Down
DROP TABLE IF EXISTS invoice_sessions;
