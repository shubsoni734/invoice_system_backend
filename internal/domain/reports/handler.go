package reports

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	reportsdb "github.com/your-org/invoice-backend/internal/domain/reports/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *reportsdb.Queries
}

func NewHandler(q *reportsdb.Queries) *Handler {
	return &Handler{q: q}
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxOrgID))
}

// GetDailyReport GET /api/v1/reports/daily?date=2025-03-17
func (h *Handler) GetDailyReport(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	report, err := h.q.GetDailyReport(context.Background(), reportsdb.GetDailyReportParams{
		OrganisationID: orgID,
		Column2:        pgtype.Date{Time: t, Valid: true},
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate daily report")
		return
	}

	response.Success(c, http.StatusOK, "Daily report retrieved", gin.H{
		"date":   dateStr,
		"report": report,
	})
}

// GetMonthlyReport GET /api/v1/reports/monthly?year=2025&month=3
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		if parsed, e := strconv.Atoi(y); e == nil {
			year = parsed
		}
	}
	if m := c.Query("month"); m != "" {
		if parsed, e := strconv.Atoi(m); e == nil {
			month = parsed
		}
	}

	// sqlc used EXTRACT on issued_date column so params are pgtype.Date holding year/month values
	// We pass the first day of the month; EXTRACT(YEAR) and EXTRACT(MONTH) will match correctly
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	report, err := h.q.GetMonthlyReport(context.Background(), reportsdb.GetMonthlyReportParams{
		OrganisationID: orgID,
		Column2:        pgtype.Date{Time: firstOfMonth, Valid: true},
		Column3:        pgtype.Date{Time: lastOfMonth, Valid: true},
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate monthly report")
		return
	}

	response.Success(c, http.StatusOK, "Monthly report retrieved", gin.H{
		"year":   year,
		"month":  month,
		"report": report,
	})
}

// GetCustomerReport GET /api/v1/reports/customer/:id
func (h *Handler) GetCustomerReport(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	invoices, err := h.q.GetCustomerInvoiceHistory(context.Background(), reportsdb.GetCustomerInvoiceHistoryParams{
		OrganisationID: orgID,
		CustomerID:     customerID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch customer invoice history")
		return
	}
	if invoices == nil {
		invoices = []reportsdb.GetCustomerInvoiceHistoryRow{}
	}

	response.Success(c, http.StatusOK, "Customer invoice history retrieved", gin.H{
		"customer_id": customerID,
		"invoices":    invoices,
	})
}

// GetRevenueSummary GET /api/v1/reports/revenue
func (h *Handler) GetRevenueSummary(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	summary, err := h.q.GetRevenueSummary(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch revenue summary")
		return
	}
	if summary == nil {
		summary = []reportsdb.GetRevenueSummaryRow{}
	}

	response.Success(c, http.StatusOK, "Revenue summary retrieved", summary)
}
