package invoices

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q  *invoicesdb.Queries
	db *pgx.Conn // Using pgx.Conn or pgxpool.Pool would be necessary for transactions, but we can manage without for now.
}

func NewHandler(q *invoicesdb.Queries) *Handler {
	return &Handler{q: q}
}

type InvoiceItemRequest struct {
	ServiceID   *uuid.UUID `json:"service_id"`
	Description string     `json:"description" binding:"required"`
	Quantity    float64    `json:"quantity" binding:"required"`
	UnitPrice   float64    `json:"unit_price" binding:"required"`
	TaxRate     float64    `json:"tax_rate"`
	SortOrder   int32      `json:"sort_order"`
}

type CreateInvoiceRequest struct {
	CustomerID    uuid.UUID            `json:"customer_id" binding:"required"`
	SessionID     uuid.UUID            `json:"session_id" binding:"required"`
	InvoiceNumber string               `json:"invoice_number" binding:"required"` // Can be generated, but taking from client for simplicity now
	IssuedDate    string               `json:"issued_date" binding:"required"`
	DueDate       string               `json:"due_date" binding:"required"`
	Currency      string               `json:"currency"`
	Notes         *string              `json:"notes"`
	Terms         *string              `json:"terms"`
	TemplateID    *uuid.UUID           `json:"template_id"`
	Items         []InvoiceItemRequest `json:"items" binding:"required,min=1"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDRaw := c.GetString(constants.CtxUserID)
	return uuid.Parse(userIDRaw)
}

func floatToNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	strValue := fmt.Sprintf("%v", f)
	err := num.Scan(strValue)
	if err != nil {
		fmt.Printf("floatToNumeric conversion error for %v: %v\n", f, err)
		num.Valid = false
	} else {
		num.Valid = true
	}
	return num
}

// GetInvoices GET /api/v1/invoices
func (h *Handler) GetInvoices(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	res, err := h.q.GetInvoices(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoices")
		return
	}

	if res == nil {
		res = []invoicesdb.Invoice{}
	}

	response.Success(c, http.StatusOK, "Invoices retrieved successfully", res)
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

	issuedDate, err := time.Parse("2006-01-02", req.IssuedDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid issued_date format (YYYY-MM-DD)")
		return
	}
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid due_date format (YYYY-MM-DD)")
		return
	}

	// Calculate totals based on items
	var totalSubtotal, totalTax, totalDiscount, grandTotal float64

	// Let's create the invoice first. We should ideally use a transaction for invoices + items.
	// For simplicity, we are calculating and executing here directly.
	for _, item := range req.Items {
		lineTotal := item.Quantity * item.UnitPrice
		taxAmount := lineTotal * (item.TaxRate / 100)
		totalSubtotal += lineTotal
		totalTax += taxAmount
		grandTotal += (lineTotal + taxAmount)
	}

	currency := "USD"
	if req.Currency != "" {
		currency = req.Currency
	}

	var tempID pgtype.UUID
	if req.TemplateID != nil {
		tempID = pgtype.UUID{Bytes: *req.TemplateID, Valid: true}
	}

	createdBy := pgtype.UUID{Bytes: userID, Valid: true}

	inv, err := h.q.CreateInvoice(context.Background(), invoicesdb.CreateInvoiceParams{
		OrganisationID: orgID,
		CustomerID:     req.CustomerID,
		SessionID:      req.SessionID,
		InvoiceNumber:  req.InvoiceNumber,
		Status:         "draft",
		IssuedDate:     pgtype.Date{Time: issuedDate, Valid: true},
		DueDate:        pgtype.Date{Time: dueDate, Valid: true},
		Subtotal:       floatToNumeric(totalSubtotal),
		TaxAmount:      floatToNumeric(totalTax),
		DiscountAmount: floatToNumeric(totalDiscount),
		Total:          floatToNumeric(grandTotal),
		Currency:       currency,
		Notes:          req.Notes,
		Terms:          req.Terms,
		TemplateID:     tempID,
		CreatedBy:      createdBy,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create invoice: %v", err))
		return
	}

	// Insert items
	var createdItems []invoicesdb.InvoiceItem
	for _, item := range req.Items {
		lineTotal := item.Quantity * item.UnitPrice
		taxAmount := lineTotal * (item.TaxRate / 100)

		var servID pgtype.UUID
		if item.ServiceID != nil {
			servID = pgtype.UUID{Bytes: *item.ServiceID, Valid: true}
		}

		dbItem, err := h.q.CreateInvoiceItem(context.Background(), invoicesdb.CreateInvoiceItemParams{
			InvoiceID:   inv.ID,
			ServiceID:   servID,
			Description: item.Description,
			Quantity:    floatToNumeric(item.Quantity),
			UnitPrice:   floatToNumeric(item.UnitPrice),
			TaxRate:     floatToNumeric(item.TaxRate),
			TaxAmount:   floatToNumeric(taxAmount),
			LineTotal:   floatToNumeric(lineTotal + taxAmount),
			SortOrder:   item.SortOrder,
		})

		if err == nil {
			createdItems = append(createdItems, dbItem)
		}
	}

	response.Success(c, http.StatusCreated, "Invoice created successfully", gin.H{
		"invoice": inv,
		"items":   createdItems,
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

	inv, err := h.q.CancelInvoice(context.Background(), invoicesdb.CancelInvoiceParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to cancel invoice")
		return
	}

	response.Success(c, http.StatusOK, "Invoice cancelled successfully", inv)
}
