-- +goose Up
CREATE TABLE organisations (
    id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                      TEXT NOT NULL,
    slug                      TEXT UNIQUE NOT NULL,
    email                     TEXT,
    phone                     TEXT,
    address                   TEXT,
    logo_url                  TEXT,
    status                    VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by_super_admin_id UUID REFERENCES super_admins(id),
    password_hash             TEXT,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS organisations;
