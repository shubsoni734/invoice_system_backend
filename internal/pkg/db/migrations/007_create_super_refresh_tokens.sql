-- +goose Up
CREATE TABLE super_refresh_tokens (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    super_admin_id UUID NOT NULL REFERENCES super_admins(id) ON DELETE CASCADE,
    token_hash     TEXT NOT NULL UNIQUE,
    expires_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ,
    ip_address     INET,
    user_agent     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_srt_super_admin ON super_refresh_tokens(super_admin_id);

-- +goose Down
DROP TABLE IF EXISTS super_refresh_tokens;
