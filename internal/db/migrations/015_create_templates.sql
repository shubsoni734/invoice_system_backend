-- +goose Up
CREATE TABLE templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    html_content    TEXT NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    thumbnail_url   TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_templates_org ON templates(organisation_id);

-- +goose Down
DROP TABLE IF EXISTS templates;
