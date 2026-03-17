package pdf

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	db          *pgxpool.Pool
	invoiceQ    *invoicesdb.Queries
	customerQ   *customersdb.Queries
	settingsQ   *settingsdb.Queries
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		db:        db,
		invoiceQ:  invoicesdb.New(db),
		customerQ: customersdb.New(db),
		settingsQ: settingsdb.New(db),
	}
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxOrgID))
}

// GenerateInvoicePDF GET /api/v1/invoices/:id/pdf
func (h *Handler) GenerateInvoicePDF(c *gin.Context) {
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

	ctx := context.Background()

	// Fetch invoice
	inv, err := h.invoiceQ.GetInvoiceByID(ctx, invoicesdb.GetInvoiceByIDParams{
		ID:             invoiceID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}

	// Fetch items
	items, err := h.invoiceQ.GetInvoiceItems(ctx, inv.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoice items")
		return
	}

	// Fetch customer
	customer, err := h.customerQ.GetCustomerByID(ctx, customersdb.GetCustomerByIDParams{
		ID:             inv.CustomerID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch customer")
		return
	}

	// Fetch settings (best-effort — use zero value if not found)
	settings, _ := h.settingsQ.GetSettings(ctx, orgID)

	// Build PDF
	pdfBytes, err := buildInvoicePDF(inv, items, customer, settings)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("invoice-%s.pdf", inv.InvoiceNumber)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
