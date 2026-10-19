-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT;

ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_api_key TEXT;
ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_sender_email TEXT;
ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_sender_name TEXT;
ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_smtp_host TEXT NOT NULL DEFAULT 'smtp-relay.brevo.com';
ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_smtp_port TEXT NOT NULL DEFAULT '587';
ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_smtp_login TEXT;
ALTER TABLE settings ADD COLUMN IF NOT EXISTS brevo_smtp_key TEXT;
ALTER TABLE settings ADD COLUMN IF NOT EXISTS notify_email_enabled BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE settings ADD COLUMN IF NOT EXISTS push_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    message         TEXT NOT NULL,
    kind            TEXT NOT NULL DEFAULT 'info',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notifications_org_created ON notifications (organisation_id, created_at DESC);

CREATE TABLE IF NOT EXISTS notification_reads (
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (notification_id, user_id)
);

CREATE TABLE IF NOT EXISTS push_subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint        TEXT NOT NULL,
    p256dh          TEXT NOT NULL,
    auth_secret     TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, endpoint)
);

-- +goose Down
DROP TABLE IF EXISTS push_subscriptions;
DROP TABLE IF EXISTS notification_reads;
DROP TABLE IF EXISTS notifications;
ALTER TABLE settings DROP COLUMN IF EXISTS push_enabled;
ALTER TABLE settings DROP COLUMN IF EXISTS notify_email_enabled;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_smtp_key;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_smtp_login;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_smtp_port;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_smtp_host;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_sender_name;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_sender_email;
ALTER TABLE settings DROP COLUMN IF EXISTS brevo_api_key;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
