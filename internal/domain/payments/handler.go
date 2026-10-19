package payments

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
	"github.com/your-org/invoice-backend/internal/domain/notify"
	paymentsdb "github.com/your-org/invoice-backend/internal/domain/payments/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/email"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q  *paymentsdb.Queries
	db *pgxpool.Pool
}

func NewHandler(q *paymentsdb.Queries, db *pgxpool.Pool) *Handler {
	return &Handler{q: q, db: db}
}

type RecordPaymentRequest struct {
	InvoiceID   uuid.UUID `json:"invoice_id"`
	Amount      float64   `json:"amount"`
	Method      *string   `json:"method"`
	Reference   *string   `json:"reference"`
	Notes       *string   `json:"notes"`
	PaymentDate string    `json:"payment_date"` // YYYY-MM-DD, defaults to today
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxOrgID))
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxUserID))
}

// GetPaymentsByInvoice GET /api/v1/invoices/:id/payments
func (h *Handler) GetPaymentsByInvoice(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	payments, err := h.q.GetPaymentsByInvoice(context.Background(), paymentsdb.GetPaymentsByInvoiceParams{
		InvoiceID:      invoiceID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}
	if payments == nil {
		payments = []paymentsdb.Payment{}
	}

	response.Success(c, http.StatusOK, "Payments retrieved", payments)
}

// RecordPayment POST /api/v1/payments
func (h *Handler) RecordPayment(c *gin.Context) {
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

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.InvoiceID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "invoice_id is required")
		return
	}
	if req.Amount <= 0 {
		response.Error(c, http.StatusBadRequest, "amount must be greater than 0")
		return
	}

	ctx := context.Background()

	summary, err := h.q.GetInvoicePaymentSummary(ctx, paymentsdb.GetInvoicePaymentSummaryParams{
		ID:             req.InvoiceID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	if summary.Status == "cancelled" {
		response.Error(c, http.StatusBadRequest, "Cannot record a payment on a cancelled invoice")
		return
	}

	balanceDue := summary.GrandTotal - summary.TotalPaid
	if balanceDue <= 0.009 {
		response.Error(c, http.StatusBadRequest, "This invoice is already fully paid")
		return
	}
	if req.Amount > balanceDue+0.009 {
		response.Error(c, http.StatusBadRequest, "Payment amount is greater than the balance due")
		return
	}

	payDate := time.Now()
	if req.PaymentDate != "" {
		if parsed, parseErr := time.Parse("2006-01-02", req.PaymentDate); parseErr == nil {
			payDate = parsed
		}
	}

	payment, err := h.q.CreatePayment(ctx, paymentsdb.CreatePaymentParams{
		OrganisationID: orgID,
		InvoiceID:      req.InvoiceID,
		Amount:         req.Amount,
		Method:         req.Method,
		Reference:      req.Reference,
		Notes:          req.Notes,
		PaymentDate:    pgtype.Date{Time: payDate, Valid: true},
		RecordedBy:     pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to record payment")
		return
	}

	updated, summaryErr := h.q.GetInvoicePaymentSummary(ctx, paymentsdb.GetInvoicePaymentSummaryParams{
		ID:             req.InvoiceID,
		OrganisationID: orgID,
	})
	balanceAfter := balanceDue - req.Amount
	if summaryErr == nil {
		balanceAfter = updated.GrandTotal - updated.TotalPaid
		if updated.GrandTotal > 0 && updated.TotalPaid+0.009 >= updated.GrandTotal {
			_ = h.q.MarkInvoicePaid(ctx, paymentsdb.MarkInvoicePaidParams{
				ID:             req.InvoiceID,
				OrganisationID: orgID,
			})
			balanceAfter = 0
		}
	}

	notify.Publish(orgID, "Payment received", fmt.Sprintf("A payment of %.2f was recorded.", req.Amount), "success")

	emailSent, emailTo, emailNote := h.emailPaymentReceipt(c, orgID, req.InvoiceID, req.Amount, balanceAfter, payDate, req.Method, req.Reference)

	response.Success(c, http.StatusCreated, "Payment recorded successfully", gin.H{
		"payment":     payment,
		"total_paid":  summary.TotalPaid + req.Amount,
		"balance_due": balanceAfter,
		"email_sent":  emailSent,
		"email_to":    emailTo,
		"email_note":  emailNote,
	})
}

func (h *Handler) emailPaymentReceipt(c *gin.Context, orgID, invoiceID uuid.UUID, amount, balance float64, payDate time.Time, method, reference *string) (bool, string, string) {
	var invoiceNumber, currency, customerName, customerEmail, businessName, businessEmail string
	err := h.db.QueryRow(c.Request.Context(), `
		SELECT i.invoice_number,
		       COALESCE(NULLIF(i.currency, ''), 'INR'),
		       COALESCE(NULLIF(c.company_name, ''), NULLIF(c.name, ''), 'Customer'),
		       COALESCE(c.email, ''),
		       COALESCE(NULLIF(s.business_name, ''), NULLIF(o.name, ''), 'Invoice Flow'),
		       COALESCE(s.business_email, '')
		FROM invoices i
		JOIN customers c ON c.id = i.customer_id
		JOIN organisations o ON o.id = i.organisation_id
		LEFT JOIN settings s ON s.organisation_id = i.organisation_id
		WHERE i.id = $1 AND i.organisation_id = $2`, invoiceID, orgID).Scan(
		&invoiceNumber, &currency, &customerName, &customerEmail, &businessName, &businessEmail,
	)
	if err != nil {
		return false, "", "Payment saved. The receipt email could not be prepared."
	}
	customerEmail = strings.TrimSpace(customerEmail)
	if !strings.Contains(customerEmail, "@") {
		return false, "", "Payment saved. This customer has no email address, so a receipt was not sent."
	}
	client := notify.MailClientForOrg(c)
	if client == nil || !client.Configured() {
		return false, customerEmail, "Payment saved. Email is not configured, so a receipt was not sent."
	}
	ref := ""
	if reference != nil {
		ref = strings.TrimSpace(*reference)
	}
	sendErr := client.SendPaymentEmail(customerEmail, email.PaymentEmailData{
		BusinessName:  businessName,
		BusinessEmail: strings.TrimSpace(businessEmail),
		CustomerName:  customerName,
		InvoiceNumber: invoiceNumber,
		Currency:      currency,
		Amount:        fmt.Sprintf("%.2f", amount),
		Method:        paymentMethodLabel(method),
		PaymentDate:   payDate.Format("02 Jan 2006"),
		Reference:     ref,
		BalanceDue:    fmt.Sprintf("%.2f", balance),
		FullyPaid:     balance <= 0.009,
	})
	if sendErr != nil {
		return false, customerEmail, "Payment saved, but the receipt email could not be sent."
	}
	return true, customerEmail, fmt.Sprintf("A receipt was emailed to %s.", customerEmail)
}

func paymentMethodLabel(method *string) string {
	if method == nil || strings.TrimSpace(*method) == "" {
		return "Payment"
	}
	switch strings.ToLower(strings.TrimSpace(*method)) {
	case "cash":
		return "Cash"
	case "upi":
		return "UPI"
	case "bank_transfer":
		return "Bank transfer"
	case "cheque":
		return "Cheque"
	case "online":
		return "Online"
	default:
		return strings.TrimSpace(*method)
	}
}

// ListPayments GET /api/v1/payments
func (h *Handler) ListPayments(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	payments, err := h.q.GetAllPaymentsWithDetails(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}
	if payments == nil {
		payments = []paymentsdb.PaymentWithDetailsRow{}
	}

	response.Success(c, http.StatusOK, "Payments retrieved", payments)
}

type UpdatePaymentRequest struct {
	Amount      float64 `json:"amount"`
	Method      *string `json:"method"`
	Reference   *string `json:"reference"`
	Notes       *string `json:"notes"`
	PaymentDate string  `json:"payment_date"`
}

// UpdatePayment PUT /api/v1/payments/:id
func (h *Handler) UpdatePayment(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}
	var req UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Amount <= 0 {
		response.Error(c, http.StatusBadRequest, "amount must be greater than 0")
		return
	}

	ctx := context.Background()
	existing, err := h.q.GetPaymentByID(ctx, paymentsdb.GetPaymentByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Payment not found")
		return
	}
	summary, err := h.q.GetInvoicePaymentSummary(ctx, paymentsdb.GetInvoicePaymentSummaryParams{
		ID:             existing.InvoiceID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	if summary.Status == "cancelled" {
		response.Error(c, http.StatusBadRequest, "Cannot update a payment on a cancelled invoice")
		return
	}
	otherPaid := summary.TotalPaid - existing.Amount
	if req.Amount > summary.GrandTotal-otherPaid+0.009 {
		response.Error(c, http.StatusBadRequest, "Payment amount is greater than the balance due")
		return
	}

	payDate := existing.PaymentDate
	if req.PaymentDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.PaymentDate)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid payment_date format (YYYY-MM-DD)")
			return
		}
		payDate = pgtype.Date{Time: parsed, Valid: true}
	}
	method := existing.Method
	if req.Method != nil && stringsTrim(*req.Method) != "" {
		method = req.Method
	}

	_, err = h.db.Exec(ctx, `
		UPDATE payments
		SET amount = $3, method = $4, reference = $5, notes = $6, payment_date = $7
		WHERE id = $1 AND organisation_id = $2
	`, id, orgID, req.Amount, method, req.Reference, req.Notes, payDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update payment")
		return
	}
	h.syncInvoicePaid(ctx, orgID, existing.InvoiceID)
	notify.Publish(orgID, "Payment updated", fmt.Sprintf("A payment was updated to %.2f.", req.Amount), "info")

	updated, _ := h.q.GetPaymentByID(ctx, paymentsdb.GetPaymentByIDParams{ID: id, OrganisationID: orgID})
	response.Success(c, http.StatusOK, "Payment updated", updated)
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

func (h *Handler) syncInvoicePaid(ctx context.Context, orgID, invoiceID uuid.UUID) {
	summary, err := h.q.GetInvoicePaymentSummary(ctx, paymentsdb.GetInvoicePaymentSummaryParams{
		ID:             invoiceID,
		OrganisationID: orgID,
	})
	if err != nil || summary.Status == "cancelled" {
		return
	}
	if summary.GrandTotal > 0 && summary.TotalPaid+0.009 >= summary.GrandTotal {
		_ = h.q.MarkInvoicePaid(ctx, paymentsdb.MarkInvoicePaidParams{ID: invoiceID, OrganisationID: orgID})
		return
	}
	if summary.Status == "paid" {
		_ = h.q.ReverseInvoicePaidStatus(ctx, paymentsdb.ReverseInvoicePaidStatusParams{ID: invoiceID, OrganisationID: orgID})
	}
}

// DeletePayment DELETE /api/v1/payments/:id
func (h *Handler) DeletePayment(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	ctx := context.Background()

	// Fetch payment to get invoice_id before deleting
	existing, err := h.q.GetPaymentByID(ctx, paymentsdb.GetPaymentByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Payment not found")
		return
	}

	if err := h.q.DeletePayment(ctx, paymentsdb.DeletePaymentParams{
		ID:             id,
		OrganisationID: orgID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete payment")
		return
	}

	// Recalculate invoice status: if it was paid and total_paid < grand_total, revert to 'sent'
	summary, summaryErr := h.q.GetInvoicePaymentSummary(ctx, paymentsdb.GetInvoicePaymentSummaryParams{
		ID:             existing.InvoiceID,
		OrganisationID: orgID,
	})
	if summaryErr == nil && summary.Status == "paid" && summary.TotalPaid < summary.GrandTotal {
		_ = h.q.ReverseInvoicePaidStatus(ctx, paymentsdb.ReverseInvoicePaidStatusParams{
			ID:             existing.InvoiceID,
			OrganisationID: orgID,
		})
	}

	notify.Publish(orgID, "Payment removed", fmt.Sprintf("A payment of %.2f was removed.", existing.Amount), "warning")
	response.Success(c, http.StatusOK, "Payment deleted", nil)
}
