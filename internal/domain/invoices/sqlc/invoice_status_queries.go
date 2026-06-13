package invoicesdb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const setInvoiceSent = `
UPDATE invoices
SET status = 'sent', sent_at = NOW(), updated_at = NOW()
WHERE id = $1 AND organisation_id = $2 AND status = 'draft'
RETURNING id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at
`

type SetInvoiceSentParams struct {
	ID             uuid.UUID `json:"id"`
	OrganisationID uuid.UUID `json:"organisation_id"`
}

func (q *Queries) SetInvoiceSent(ctx context.Context, arg SetInvoiceSentParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, setInvoiceSent, arg.ID, arg.OrganisationID)
	var i Invoice
	err := row.Scan(
		&i.ID, &i.OrganisationID, &i.CustomerID, &i.SessionID, &i.InvoiceNumber, &i.Status,
		&i.IssuedDate, &i.DueDate, &i.Subtotal, &i.TaxAmount, &i.CgstAmount, &i.SgstAmount, &i.IgstAmount,
		&i.DiscountAmount, &i.Total, &i.GrandTotal, &i.IsSameState, &i.Currency, &i.Notes, &i.Terms,
		&i.TemplateID, &i.CreatedBy, &i.SentAt, &i.PaidAt, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}

const setInvoicePaid = `
UPDATE invoices
SET status = 'paid', paid_at = NOW(), updated_at = NOW()
WHERE id = $1 AND organisation_id = $2 AND status NOT IN ('cancelled')
RETURNING id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at
`

type SetInvoicePaidParams struct {
	ID             uuid.UUID `json:"id"`
	OrganisationID uuid.UUID `json:"organisation_id"`
}

func (q *Queries) SetInvoicePaid(ctx context.Context, arg SetInvoicePaidParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, setInvoicePaid, arg.ID, arg.OrganisationID)
	var i Invoice
	err := row.Scan(
		&i.ID, &i.OrganisationID, &i.CustomerID, &i.SessionID, &i.InvoiceNumber, &i.Status,
		&i.IssuedDate, &i.DueDate, &i.Subtotal, &i.TaxAmount, &i.CgstAmount, &i.SgstAmount, &i.IgstAmount,
		&i.DiscountAmount, &i.Total, &i.GrandTotal, &i.IsSameState, &i.Currency, &i.Notes, &i.Terms,
		&i.TemplateID, &i.CreatedBy, &i.SentAt, &i.PaidAt, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}

const updateInvoiceFields = `
UPDATE invoices
SET notes = $3,
    terms = $4,
    due_date = CASE WHEN $5::DATE IS NULL THEN due_date ELSE $5::DATE END,
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING id, organisation_id, customer_id, session_id, invoice_number, status,
    issued_date, due_date, subtotal, tax_amount, cgst_amount, sgst_amount, igst_amount,
    discount_amount, total, grand_total, is_same_state, currency, notes, terms,
    template_id, created_by, sent_at, paid_at, created_at, updated_at
`

type UpdateInvoiceFieldsParams struct {
	ID             uuid.UUID   `json:"id"`
	OrganisationID uuid.UUID   `json:"organisation_id"`
	Notes          *string     `json:"notes"`
	Terms          *string     `json:"terms"`
	DueDate        pgtype.Date `json:"due_date"`
}

func (q *Queries) UpdateInvoiceFields(ctx context.Context, arg UpdateInvoiceFieldsParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, updateInvoiceFields,
		arg.ID, arg.OrganisationID, arg.Notes, arg.Terms, arg.DueDate)
	var i Invoice
	err := row.Scan(
		&i.ID, &i.OrganisationID, &i.CustomerID, &i.SessionID, &i.InvoiceNumber, &i.Status,
		&i.IssuedDate, &i.DueDate, &i.Subtotal, &i.TaxAmount, &i.CgstAmount, &i.SgstAmount, &i.IgstAmount,
		&i.DiscountAmount, &i.Total, &i.GrandTotal, &i.IsSameState, &i.Currency, &i.Notes, &i.Terms,
		&i.TemplateID, &i.CreatedBy, &i.SentAt, &i.PaidAt, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}
