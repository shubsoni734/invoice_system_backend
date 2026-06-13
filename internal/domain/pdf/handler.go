package pdf

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	templatesdb "github.com/your-org/invoice-backend/internal/domain/templates/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	db         *pgxpool.Pool
	invoiceQ   *invoicesdb.Queries
	customerQ  *customersdb.Queries
	settingsQ  *settingsdb.Queries
	templateQ  *templatesdb.Queries
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		db:        db,
		invoiceQ:  invoicesdb.New(db),
		customerQ: customersdb.New(db),
		settingsQ: settingsdb.New(db),
		templateQ: templatesdb.New(db),
	}
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxOrgID))
}

// GenerateInvoicePDF GET /api/v1/invoices/:id/pdf
// Returns a rendered HTML page with auto-print so the user can save as PDF.
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

	inv, err := h.invoiceQ.GetInvoiceByID(ctx, invoicesdb.GetInvoiceByIDParams{
		ID:             invoiceID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}

	items, err := h.invoiceQ.GetInvoiceItems(ctx, inv.ID)
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

	// Try invoice-level template, then org default template
	var customTemplateHTML string
	if templateID, ok := resolveTemplateID(inv, settings); ok {
		if t, err := h.templateQ.GetTemplateByID(ctx, templatesdb.GetTemplateByIDParams{
			ID:             templateID,
			OrganisationID: orgID,
		}); err == nil {
			customTemplateHTML = t.HtmlContent
		}
	}

	html := buildInvoiceHTML(inv, items, customer, settings, customTemplateHTML)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", `inline; filename="invoice.html"`)
	c.String(http.StatusOK, html)
}

// resolveTemplateID returns the template UUID to use for this invoice.
func resolveTemplateID(inv invoicesdb.Invoice, settings settingsdb.Setting) (uuid.UUID, bool) {
	if inv.TemplateID.Valid {
		return inv.TemplateID.Bytes, true
	}
	if settings.DefaultTemplateID.Valid {
		return settings.DefaultTemplateID.Bytes, true
	}
	return uuid.UUID{}, false
}
