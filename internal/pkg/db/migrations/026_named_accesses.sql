-- +goose Up
DROP TABLE IF EXISTS accesses;

CREATE TABLE accesses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    resource        TEXT NOT NULL,
    can_create      BOOLEAN NOT NULL DEFAULT FALSE,
    can_read        BOOLEAN NOT NULL DEFAULT FALSE,
    can_update      BOOLEAN NOT NULL DEFAULT FALSE,
    can_delete      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organisation_id, name)
);

CREATE TABLE role_accesses (
    role_id   UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    access_id UUID NOT NULL REFERENCES accesses(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, access_id)
);

-- +goose Down
DROP TABLE IF EXISTS role_accesses;
DROP TABLE IF EXISTS accesses;
