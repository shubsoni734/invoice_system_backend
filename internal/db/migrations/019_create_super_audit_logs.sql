-- +goose Up
CREATE TABLE super_audit_logs (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    super_admin_id UUID REFERENCES super_admins(id),
    action         TEXT NOT NULL,
    target_type    TEXT,
    target_id      UUID,
    details        JSONB,
    ip_address     INET,
    user_agent     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_super_audit ON super_audit_logs(super_admin_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS super_audit_logs;
