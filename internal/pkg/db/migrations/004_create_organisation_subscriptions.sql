-- +goose Up
CREATE TABLE organisation_subscriptions (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id      UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    plan_id              UUID NOT NULL REFERENCES plans(id),
    status               TEXT NOT NULL DEFAULT 'trialing',
    trial_ends_at        TIMESTAMPTZ,
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end   TIMESTAMPTZ NOT NULL,
    cancelled_at         TIMESTAMPTZ,
    external_id          TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sub_org ON organisation_subscriptions(organisation_id);

-- +goose Down
DROP TABLE IF EXISTS organisation_subscriptions;
