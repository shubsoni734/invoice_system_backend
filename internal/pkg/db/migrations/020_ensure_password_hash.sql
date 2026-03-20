-- +goose Up
ALTER TABLE organisations ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- +goose Down
ALTER TABLE organisations DROP COLUMN IF EXISTS password_hash;
