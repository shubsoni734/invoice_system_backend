-- +goose Up

-- 1. Add new columns
ALTER TABLE services
    ADD COLUMN type               VARCHAR(20)   NOT NULL DEFAULT 'service',
    ADD COLUMN hsn_code           VARCHAR(20),
    ADD COLUMN sac_code           VARCHAR(20),
    ADD COLUMN tax_preference     VARCHAR(20)   NOT NULL DEFAULT 'taxable',
    ADD COLUMN intra_state_tax_rate NUMERIC(5,2) NOT NULL DEFAULT 18,
    ADD COLUMN inter_state_tax_rate NUMERIC(5,2) NOT NULL DEFAULT 18,
    ADD COLUMN selling_price      NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN account            VARCHAR(150);

-- 2. Migrate existing data
UPDATE services
SET
    selling_price        = unit_price,
    intra_state_tax_rate = tax_rate,
    inter_state_tax_rate = tax_rate;

-- 3. Drop old columns
ALTER TABLE services
    DROP COLUMN unit_price,
    DROP COLUMN tax_rate;

-- 4. Make unit nullable (was TEXT NOT NULL DEFAULT 'unit')
ALTER TABLE services
    ALTER COLUMN unit DROP NOT NULL,
    ALTER COLUMN unit DROP DEFAULT,
    ALTER COLUMN unit TYPE VARCHAR(50);

-- 5. Add indexes
CREATE INDEX IF NOT EXISTS idx_services_type ON services(type);
CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);

-- +goose Down

-- Re-add old columns
ALTER TABLE services
    ADD COLUMN unit_price NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN tax_rate   NUMERIC(5,2)  NOT NULL DEFAULT 0;

-- Restore data
UPDATE services
SET
    unit_price = selling_price,
    tax_rate   = intra_state_tax_rate;

-- Drop new columns
ALTER TABLE services
    DROP COLUMN type,
    DROP COLUMN hsn_code,
    DROP COLUMN sac_code,
    DROP COLUMN tax_preference,
    DROP COLUMN intra_state_tax_rate,
    DROP COLUMN inter_state_tax_rate,
    DROP COLUMN selling_price,
    DROP COLUMN account;

-- Restore unit to NOT NULL
ALTER TABLE services
    ALTER COLUMN unit SET NOT NULL,
    ALTER COLUMN unit SET DEFAULT 'unit',
    ALTER COLUMN unit TYPE TEXT;

DROP INDEX IF EXISTS idx_services_type;
DROP INDEX IF EXISTS idx_services_name;
