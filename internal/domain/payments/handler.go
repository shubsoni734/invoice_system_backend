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
	InvoiceID   uuid.UUID `json:"invoice_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Method      string    `json:"method" binding:"required"`
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

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Scan(f)
	return n
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

	payDate := time.Now()
	if req.PaymentDate != "" {
		if parsed, parseErr := time.Parse("2006-01-02", req.PaymentDate); parseErr == nil {
			payDate = parsed
		}
	}

	method := req.Method
	if method == "" {
		method = constants.PaymentCash
	}

	payment, err := h.q.CreatePayment(context.Background(), paymentsdb.CreatePaymentParams{
		OrganisationID: orgID,
		InvoiceID:      req.InvoiceID,
		Amount:         floatToNumeric(req.Amount),
		Method:         method,
		Reference:      req.Reference,
		Notes:          req.Notes,
		PaymentDate:    pgtype.Date{Time: payDate, Valid: true},
		RecordedBy:     pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to record payment")
		return
	}

	response.Success(c, http.StatusCreated, "Payment recorded successfully", payment)
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

	if err := h.q.DeletePayment(context.Background(), paymentsdb.DeletePaymentParams{
		ID:             id,
		OrganisationID: orgID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete payment")
		return
	}

	response.Success(c, http.StatusOK, "Payment deleted", nil)
}
