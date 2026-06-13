-- +goose Up

-- Add GST split columns and grand_total to invoices
ALTER TABLE invoices
    ADD COLUMN cgst_amount    NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN sgst_amount    NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN igst_amount    NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN grand_total    NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN is_same_state  BOOLEAN       NOT NULL DEFAULT TRUE;

-- Make session_id nullable (invoices can exist without a session)
ALTER TABLE invoices ALTER COLUMN session_id DROP NOT NULL;

-- Add HSN/SAC code to invoice items
ALTER TABLE invoice_items
    ADD COLUMN hsn_sac_code VARCHAR(20);

-- Make payment method optional
ALTER TABLE payments
    ALTER COLUMN method DROP NOT NULL;

-- +goose Down

ALTER TABLE invoices
    DROP COLUMN cgst_amount,
    DROP COLUMN sgst_amount,
    DROP COLUMN igst_amount,
    DROP COLUMN grand_total,
    DROP COLUMN is_same_state;

ALTER TABLE invoices ALTER COLUMN session_id SET NOT NULL;

ALTER TABLE invoice_items
    DROP COLUMN hsn_sac_code;

ALTER TABLE payments
    ALTER COLUMN method SET NOT NULL;
