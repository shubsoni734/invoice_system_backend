package invoices

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/email"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q           *invoicesdb.Queries
	db          *pgxpool.Pool
	emailClient *email.Client
	customerQ   *customersdb.Queries
	settingsQ   *settingsdb.Queries
}

func NewHandler(q *invoicesdb.Queries, db *pgxpool.Pool, emailClient *email.Client) *Handler {
	return &Handler{
		q:           q,
		db:          db,
		emailClient: emailClient,
		customerQ:   customersdb.New(db),
		settingsQ:   settingsdb.New(db),
	}
}

type InvoiceItemRequest struct {
	ServiceID   *uuid.UUID `json:"service_id"`
	Description string     `json:"description"`
	HsnSacCode  *string    `json:"hsn_sac_code"`
	Quantity    float64    `json:"quantity"`
	UnitPrice   float64    `json:"unit_price"`
	TaxRate     float64    `json:"tax_rate"`
	SortOrder   int32      `json:"sort_order"`
}

type CreateInvoiceRequest struct {
	CustomerID     uuid.UUID            `json:"customer_id"`
	SessionID      *uuid.UUID           `json:"session_id"`
	InvoiceNumber  *string              `json:"invoice_number"`
	IssuedDate     string               `json:"issued_date"`
	DueDate        string               `json:"due_date"`
	IsSameState    *bool                `json:"is_same_state"`
	DiscountAmount *float64             `json:"discount_amount"`
	Currency       *string              `json:"currency"`
	Notes          *string              `json:"notes"`
	Terms          *string              `json:"terms"`
	TemplateID     *uuid.UUID           `json:"template_id"`
	Items          []InvoiceItemRequest `json:"items"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxOrgID))
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxUserID))
}

func parseOptionalDate(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{Valid: false}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{Valid: false}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// GetInvoices GET /api/v1/invoices
func (h *Handler) GetInvoices(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	p := utils.GetPaginationParams(c)
	status := c.Query("status")

	var customerID pgtype.UUID
	if raw := c.Query("customer_id"); raw != "" {
		if parsed, parseErr := uuid.Parse(raw); parseErr == nil {
			customerID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	fromDate, _ := parseOptionalDate(c.Query("from_date"))
	toDate, _ := parseOptionalDate(c.Query("to_date"))

	filterParams := invoicesdb.CountInvoicesParams{
		OrganisationID: orgID,
		Status:         status,
		CustomerID:     customerID,
		FromDate:       fromDate,
		ToDate:         toDate,
	}

	total, err := h.q.CountInvoices(context.Background(), filterParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to count invoices")
		return
	}

	res, err := h.q.GetInvoices(context.Background(), invoicesdb.GetInvoicesParams{
		OrganisationID: orgID,
		Status:         status,
		CustomerID:     customerID,
		FromDate:       fromDate,
		ToDate:         toDate,
		Offset:         int32(p.Offset),
		PerPage:        int32(p.PerPage),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoices")
		return
	}
	if res == nil {
		res = []invoicesdb.InvoiceRow{}
	}

	response.SuccessWithMeta(c, http.StatusOK, "Invoices retrieved successfully", res, &response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: utils.CalculateTotalPages(total, p.PerPage),
	})
}

// GetInvoiceByID GET /api/v1/invoices/:id
func (h *Handler) GetInvoiceByID(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	inv, err := h.q.GetInvoiceByID(context.Background(), invoicesdb.GetInvoiceByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}

	items, err := h.q.GetInvoiceItems(context.Background(), inv.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoice items")
		return
	}
	if items == nil {
		items = []invoicesdb.InvoiceItem{}
	}

	response.Success(c, http.StatusOK, "Invoice retrieved successfully", gin.H{
		"invoice": inv,
		"items":   items,
	})
}

// CreateInvoice POST /api/v1/invoices
func (h *Handler) CreateInvoice(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CustomerID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "customer_id is required")
		return
	}
	if len(req.Items) == 0 {
		response.Error(c, http.StatusBadRequest, "at least one item is required")
		return
	}
	for i, item := range req.Items {
		if strings.TrimSpace(item.Description) == "" {
			response.Error(c, http.StatusBadRequest, "item description is required")
			return
		}
		if item.Quantity <= 0 {
			response.Error(c, http.StatusBadRequest, "item quantity must be > 0")
			return
		}
		if item.UnitPrice < 0 {
			response.Error(c, http.StatusBadRequest, "item unit_price must be >= 0")
			return
		}
		_ = i
	}

	issuedDate, err := time.Parse("2006-01-02", req.IssuedDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid issued_date format (YYYY-MM-DD)")
		return
	}
	var dueDate time.Time
	if req.DueDate != "" {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid due_date format (YYYY-MM-DD)")
			return
		}
	} else {
		dueDate = issuedDate.AddDate(0, 0, 30)
	}

	// Resolve optional fields
	isSameState := true
	if req.IsSameState != nil {
		isSameState = *req.IsSameState
	}
	discountAmount := 0.0
	if req.DiscountAmount != nil {
		discountAmount = *req.DiscountAmount
	}
	currency := "INR"
	if req.Currency != nil && *req.Currency != "" {
		currency = *req.Currency
	}

	var templateID pgtype.UUID
	if req.TemplateID != nil {
		templateID = pgtype.UUID{Bytes: *req.TemplateID, Valid: true}
	}

	// Calculate totals
	var subtotal, totalTax float64
	for _, item := range req.Items {
		lineTotal := item.Quantity * item.UnitPrice
		taxAmt := lineTotal * (item.TaxRate / 100)
		subtotal += lineTotal
		totalTax += taxAmt
	}
	total := subtotal + totalTax - discountAmount

	var cgst, sgst, igst float64
	if isSameState {
		cgst = totalTax / 2
		sgst = totalTax / 2
	} else {
		igst = totalTax
	}
	grandTotal := total

	ctx := context.Background()

	// Begin transaction for atomic invoice number generation + insert
	tx, err := h.db.Begin(ctx)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.q.WithTx(tx)

	// Resolve invoice number
	var invoiceNumber string
	var sessionID pgtype.UUID

	if req.SessionID != nil {
		seq, seqErr := qtx.IncrementSessionSequence(ctx, invoicesdb.IncrementSessionSequenceParams{
			ID:             *req.SessionID,
			OrganisationID: orgID,
		})
		if seqErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid or unauthorised session_id")
			return
		}
		invoiceNumber = utils.FormatInvoiceNumber(seq.Prefix, int(seq.Year), int(seq.CurrentSequence))
		sessionID = pgtype.UUID{Bytes: *req.SessionID, Valid: true}
	} else if req.InvoiceNumber != nil && strings.TrimSpace(*req.InvoiceNumber) != "" {
		invoiceNumber = strings.TrimSpace(*req.InvoiceNumber)
	} else {
		// Auto-generate: org prefix + current year + (total invoices + 1)
		count, _ := qtx.CountInvoices(ctx, invoicesdb.CountInvoicesParams{
			OrganisationID: orgID,
			Status:         "",
			CustomerID:     pgtype.UUID{},
			FromDate:       pgtype.Date{},
			ToDate:         pgtype.Date{},
		})
		settings, _ := h.settingsQ.GetSettings(ctx, orgID)
		prefix := settings.InvoicePrefix
		if prefix == "" {
			prefix = "INV"
		}
		invoiceNumber = utils.FormatInvoiceNumber(prefix, time.Now().Year(), int(count)+1)
	}

	// Insert invoice
	invID, err := qtx.CreateInvoice(ctx, invoicesdb.CreateInvoiceParams{
		OrganisationID: orgID,
		CustomerID:     req.CustomerID,
		SessionID:      sessionID,
		InvoiceNumber:  invoiceNumber,
		Status:         "draft",
		IssuedDate:     pgtype.Date{Time: issuedDate, Valid: true},
		DueDate:        pgtype.Date{Time: dueDate, Valid: true},
		Subtotal:       subtotal,
		TaxAmount:      totalTax,
		CgstAmount:     cgst,
		SgstAmount:     sgst,
		IgstAmount:     igst,
		DiscountAmount: discountAmount,
		Total:          total,
		GrandTotal:     grandTotal,
		IsSameState:    isSameState,
		Currency:       currency,
		Notes:          req.Notes,
		Terms:          req.Terms,
		TemplateID:     templateID,
		CreatedBy:      pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create invoice")
		return
	}

	// Insert items
	for idx, item := range req.Items {
		lineTotal := item.Quantity * item.UnitPrice
		taxAmt := lineTotal * (item.TaxRate / 100)

		var servID pgtype.UUID
		if item.ServiceID != nil {
			servID = pgtype.UUID{Bytes: *item.ServiceID, Valid: true}
		}

		if err := qtx.CreateInvoiceItem(ctx, invoicesdb.CreateInvoiceItemParams{
			InvoiceID:   invID,
			ServiceID:   servID,
			Description: strings.TrimSpace(item.Description),
			HsnSacCode:  item.HsnSacCode,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			TaxAmount:   taxAmt,
			LineTotal:   lineTotal + taxAmt,
			SortOrder:   int32(idx),
		}); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to create invoice item")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Fetch full invoice and items after commit
	inv, _ := h.q.GetInvoiceByID(ctx, invoicesdb.GetInvoiceByIDParams{
		ID:             invID,
		OrganisationID: orgID,
	})
	items, _ := h.q.GetInvoiceItems(ctx, invID)
	if items == nil {
		items = []invoicesdb.InvoiceItem{}
	}

	response.Success(c, http.StatusCreated, "Invoice created successfully", gin.H{
		"invoice": inv,
		"items":   items,
	})
}

// CancelInvoice PUT /api/v1/invoices/:id/cancel
func (h *Handler) CancelInvoice(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	ctx := context.Background()

	inv, err := h.q.GetInvoiceForValidation(ctx, invoicesdb.GetInvoiceForValidationParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	if inv.Status == "paid" {
		response.Error(c, http.StatusBadRequest, "Cannot cancel a paid invoice")
		return
	}
	if inv.Status == "cancelled" {
		response.Error(c, http.StatusBadRequest, "Invoice is already cancelled")
		return
	}

	result, err := h.q.CancelInvoice(ctx, invoicesdb.CancelInvoiceParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to cancel invoice")
		return
	}

	response.Success(c, http.StatusOK, "Invoice cancelled successfully", result)
}

// SendInvoice PUT /api/v1/invoices/:id/send
func (h *Handler) SendInvoice(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	ctx := context.Background()

	inv, err := h.q.GetInvoiceForValidation(ctx, invoicesdb.GetInvoiceForValidationParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	if inv.Status != "draft" {
		response.Error(c, http.StatusBadRequest, "Only draft invoices can be sent")
		return
	}

	updated, err := h.q.SetInvoiceSent(ctx, invoicesdb.SetInvoiceSentParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send invoice")
		return
	}

	response.Success(c, http.StatusOK, "Invoice sent successfully", updated)
}

// MarkInvoiceAsPaid PUT /api/v1/invoices/:id/paid
func (h *Handler) MarkInvoiceAsPaid(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	ctx := context.Background()

	inv, err := h.q.GetInvoiceForValidation(ctx, invoicesdb.GetInvoiceForValidationParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	if inv.Status == "cancelled" {
		response.Error(c, http.StatusBadRequest, "Cannot mark a cancelled invoice as paid")
		return
	}

	updated, err := h.q.SetInvoicePaid(ctx, invoicesdb.SetInvoicePaidParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark invoice as paid")
		return
	}

	response.Success(c, http.StatusOK, "Invoice marked as paid", updated)
}

// UpdateInvoiceRequest holds updatable fields (notes, terms, due_date only).
type UpdateInvoiceRequest struct {
	Notes   *string `json:"notes"`
	Terms   *string `json:"terms"`
	DueDate string  `json:"due_date"`
}

// UpdateInvoice PUT /api/v1/invoices/:id
func (h *Handler) UpdateInvoice(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	dueDate, _ := parseOptionalDate(req.DueDate)

	updated, err := h.q.UpdateInvoiceFields(context.Background(), invoicesdb.UpdateInvoiceFieldsParams{
		ID:             id,
		OrganisationID: orgID,
		Notes:          req.Notes,
		Terms:          req.Terms,
		DueDate:        dueDate,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update invoice")
		return
	}

	response.Success(c, http.StatusOK, "Invoice updated successfully", updated)
}

// DeleteInvoice DELETE /api/v1/invoices/:id
func (h *Handler) DeleteInvoice(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	ctx := context.Background()

	inv, err := h.q.GetInvoiceForValidation(ctx, invoicesdb.GetInvoiceForValidationParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	if inv.Status == "paid" {
		response.Error(c, http.StatusBadRequest, "Cannot delete a paid invoice")
		return
	}
	if inv.Status == "sent" {
		response.Error(c, http.StatusBadRequest, "Cannot delete a sent invoice; cancel it first")
		return
	}

	if err := h.q.DeleteInvoice(ctx, invoicesdb.DeleteInvoiceParams{
		ID:             id,
		OrganisationID: orgID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete invoice")
		return
	}

	response.Success(c, http.StatusOK, "Invoice deleted successfully", nil)
}

// SendInvoiceEmail POST /api/v1/invoices/:id/email
func (h *Handler) SendInvoiceEmail(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	var req struct {
		RecipientEmail string `json:"recipient_email"`
	}
	_ = c.ShouldBindJSON(&req)

	ctx := context.Background()

	inv, err := h.q.GetInvoiceByID(ctx, invoicesdb.GetInvoiceByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}

	items, err := h.q.GetInvoiceItems(ctx, inv.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoice items")
		return
	}

	customer, err := h.customerQ.GetCustomerByID(ctx, customersdb.GetCustomerByIDParams{
		ID:             inv.CustomerID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch customer")
		return
	}

	settings, _ := h.settingsQ.GetSettings(ctx, orgID)

	// Use provided email, fall back to customer email
	toEmail := req.RecipientEmail
	if toEmail == "" {
		if customer.Email == nil || *customer.Email == "" {
			response.Error(c, http.StatusBadRequest, "No recipient email — provide recipient_email or add email to customer")
			return
		}
		toEmail = *customer.Email
	}

	issuedDate := ""
	if inv.IssuedDate.Valid {
		issuedDate = inv.IssuedDate.Time.Format("02 Jan 2006")
	}
	dueDate := ""
	if inv.DueDate.Valid {
		dueDate = inv.DueDate.Time.Format("02 Jan 2006")
	}

	// Build items HTML for email
	itemsHTML := buildEmailItemsHTML(items, inv.Currency)

	bizName := ""
	if settings.BusinessName != nil {
		bizName = *settings.BusinessName
	}
	bizEmail := ""
	if settings.BusinessEmail != nil {
		bizEmail = *settings.BusinessEmail
	}

	notes := ""
	if inv.Notes != nil {
		notes = *inv.Notes
	}

	data := email.InvoiceEmailData{
		BusinessName:  bizName,
		BusinessEmail: bizEmail,
		InvoiceNumber: inv.InvoiceNumber,
		IssuedDate:    issuedDate,
		DueDate:       dueDate,
		Currency:      inv.Currency,
		CustomerName:  customer.Name,
		Subtotal:      fmt.Sprintf("%.2f", inv.Subtotal),
		TaxAmount:     fmt.Sprintf("%.2f", inv.TaxAmount),
		GrandTotal:    fmt.Sprintf("%.2f", inv.GrandTotal),
		Notes:         notes,
		ItemsHTML:     itemsHTML,
	}

	if h.emailClient == nil {
		response.Error(c, http.StatusServiceUnavailable, "Email service not configured")
		return
	}

	if err := h.emailClient.SendInvoiceEmail(toEmail, data); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send invoice email")
		return
	}

	response.Success(c, http.StatusOK, "Invoice email sent successfully", gin.H{"sent_to": toEmail})
}

func buildEmailItemsHTML(items []invoicesdb.InvoiceItem, currency string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf(
			`<tr style="border-bottom:1px solid #e5e7eb">
<td style="padding:10px 12px;font-size:13px;color:#1f2937">%s</td>
<td style="padding:10px 12px;font-size:13px;color:#1f2937;text-align:right">%.2f</td>
<td style="padding:10px 12px;font-size:13px;color:#1f2937;text-align:right">%s %.2f</td>
<td style="padding:10px 12px;font-size:13px;color:#1f2937;text-align:right">%s %.2f</td>
</tr>`,
			item.Description, item.Quantity, currency, item.UnitPrice, currency, item.LineTotal,
		))
	}
	return sb.String()
}
