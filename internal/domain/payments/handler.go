package payments

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	paymentsdb "github.com/your-org/invoice-backend/internal/domain/payments/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *paymentsdb.Queries
}

func NewHandler(q *paymentsdb.Queries) *Handler {
	return &Handler{q: q}
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

	payDate := time.Now()
	if req.PaymentDate != "" {
		if parsed, parseErr := time.Parse("2006-01-02", req.PaymentDate); parseErr == nil {
			payDate = parsed
		}
	}

	ctx := context.Background()

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

	// Auto-mark invoice as paid if fully covered
	summary, summaryErr := h.q.GetInvoicePaymentSummary(ctx, paymentsdb.GetInvoicePaymentSummaryParams{
		ID:             req.InvoiceID,
		OrganisationID: orgID,
	})
	if summaryErr == nil && summary.GrandTotal > 0 && summary.TotalPaid >= summary.GrandTotal {
		_ = h.q.MarkInvoicePaid(ctx, paymentsdb.MarkInvoicePaidParams{
			ID:             req.InvoiceID,
			OrganisationID: orgID,
		})
	}

	response.Success(c, http.StatusCreated, "Payment recorded successfully", payment)
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

	response.Success(c, http.StatusOK, "Payment deleted", nil)
}
