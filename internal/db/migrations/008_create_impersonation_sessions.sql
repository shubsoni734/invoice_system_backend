-- +goose Up
CREATE TABLE impersonation_sessions (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    super_admin_id UUID NOT NULL REFERENCES super_admins(id),
    target_org_id  UUID NOT NULL REFERENCES organisations(id),
    target_user_id UUID NOT NULL REFERENCES users(id),
    reason         TEXT NOT NULL,
    token_hash     TEXT NOT NULL,
    started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at       TIMESTAMPTZ,
    ip_address     INET
);
CREATE INDEX idx_imp_super_admin ON impersonation_sessions(super_admin_id);

-- +goose Down
DROP TABLE IF EXISTS impersonation_sessions;
