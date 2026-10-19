-- +goose Up
CREATE TABLE accesses (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    resource    TEXT NOT NULL,
    can_create  BOOLEAN NOT NULL DEFAULT FALSE,
    can_read    BOOLEAN NOT NULL DEFAULT FALSE,
    can_update  BOOLEAN NOT NULL DEFAULT FALSE,
    can_delete  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (role_id, resource)
);

-- +goose Down
DROP TABLE IF EXISTS accesses;
