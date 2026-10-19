package reports

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	reportsdb "github.com/your-org/invoice-backend/internal/domain/reports/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q  *reportsdb.Queries
	db *pgxpool.Pool
}

func NewHandler(q *reportsdb.Queries, db *pgxpool.Pool) *Handler {
	return &Handler{q: q, db: db}
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

type monthPoint struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
	Paid    float64 `json:"paid"`
	Base    float64 `json:"base"`
	GST     float64 `json:"gst"`
}

type namedCount struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type customerAmount struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type overviewPayload struct {
	Year      int              `json:"year"`
	Revenue   float64          `json:"revenue"`
	Paid      float64          `json:"paid"`
	GST       float64          `json:"gst"`
	Pending   float64          `json:"pending"`
	Months    []monthPoint     `json:"months"`
	Status    []namedCount     `json:"status"`
	Customers []customerAmount `json:"customers"`
}

// GetOverview GET /api/v1/reports/overview?year=2026
func (h *Handler) GetOverview(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}
	year := time.Now().Year()
	if raw := c.Query("year"); raw != "" {
		if parsed, convErr := strconv.Atoi(raw); convErr == nil && parsed > 2000 && parsed < 2100 {
			year = parsed
		}
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	ctx := c.Request.Context()
	payload := overviewPayload{Year: year, Months: []monthPoint{}, Status: []namedCount{}, Customers: []customerAmount{}}

	err = h.db.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(COALESCE(NULLIF(grand_total, 0), total)) FROM invoices
				WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3 AND status <> 'cancelled'), 0)::float8,
			COALESCE((SELECT SUM(amount) FROM payments
				WHERE organisation_id = $1 AND payment_date BETWEEN $2 AND $3), 0)::float8,
			COALESCE((SELECT SUM(tax_amount) FROM invoices
				WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3 AND status <> 'cancelled'), 0)::float8,
			COALESCE((SELECT SUM(GREATEST(COALESCE(NULLIF(i.grand_total, 0), i.total) - COALESCE(paid.amount, 0), 0))
				FROM invoices i
				LEFT JOIN (SELECT invoice_id, SUM(amount) amount FROM payments GROUP BY invoice_id) paid ON paid.invoice_id = i.id
				WHERE i.organisation_id = $1 AND i.issued_date BETWEEN $2 AND $3 AND i.status <> 'cancelled'), 0)::float8
	`, orgID, from, to).Scan(&payload.Revenue, &payload.Paid, &payload.GST, &payload.Pending)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load the report")
		return
	}

	monthRows, err := h.db.Query(ctx, `
		SELECT to_char(d, 'Mon'),
			COALESCE((SELECT SUM(COALESCE(NULLIF(i.grand_total, 0), i.total)) FROM invoices i
				WHERE i.organisation_id = $1 AND i.status <> 'cancelled' AND date_trunc('month', i.issued_date) = d), 0)::float8,
			COALESCE((SELECT SUM(p.amount) FROM payments p
				WHERE p.organisation_id = $1 AND date_trunc('month', p.payment_date) = d), 0)::float8,
			COALESCE((SELECT SUM(i.subtotal) FROM invoices i
				WHERE i.organisation_id = $1 AND i.status <> 'cancelled' AND date_trunc('month', i.issued_date) = d), 0)::float8,
			COALESCE((SELECT SUM(i.tax_amount) FROM invoices i
				WHERE i.organisation_id = $1 AND i.status <> 'cancelled' AND date_trunc('month', i.issued_date) = d), 0)::float8
		FROM generate_series($2::date, $3::date, '1 month') d
		ORDER BY d
	`, orgID, from, time.Date(year, 12, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load monthly figures")
		return
	}
	defer monthRows.Close()
	for monthRows.Next() {
		var point monthPoint
		if err := monthRows.Scan(&point.Month, &point.Revenue, &point.Paid, &point.Base, &point.GST); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read monthly figures")
			return
		}
		payload.Months = append(payload.Months, point)
	}

	statusRows, err := h.db.Query(ctx, `
		SELECT
			CASE
				WHEN status = 'paid' THEN 'Paid'
				WHEN status = 'cancelled' THEN 'Cancelled'
				WHEN status = 'draft' THEN 'Draft'
				WHEN due_date < CURRENT_DATE THEN 'Overdue'
				ELSE 'Sent'
			END,
			COUNT(*)
		FROM invoices
		WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3
		GROUP BY 1
		ORDER BY 1
	`, orgID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load invoice status")
		return
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var item namedCount
		if err := statusRows.Scan(&item.Name, &item.Value); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read invoice status")
			return
		}
		payload.Status = append(payload.Status, item)
	}

	customerRows, err := h.db.Query(ctx, `
		SELECT COALESCE(NULLIF(c.company_name, ''), c.name, 'Customer'),
			COALESCE(SUM(COALESCE(NULLIF(i.grand_total, 0), i.total)), 0)::float8
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		WHERE i.organisation_id = $1 AND i.issued_date BETWEEN $2 AND $3 AND i.status <> 'cancelled'
		GROUP BY 1
		ORDER BY 2 DESC
		LIMIT 8
	`, orgID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load customer totals")
		return
	}
	defer customerRows.Close()
	for customerRows.Next() {
		var item customerAmount
		if err := customerRows.Scan(&item.Name, &item.Amount); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read customer totals")
			return
		}
		payload.Customers = append(payload.Customers, item)
	}

	response.Success(c, http.StatusOK, "Report overview", payload)
}

type gstLine struct {
	ID            uuid.UUID `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	CustomerName  string    `json:"customer_name"`
	IssuedDate    string    `json:"issued_date"`
	SameState     bool      `json:"same_state"`
	Taxable       float64   `json:"taxable"`
	CGST          float64   `json:"cgst"`
	SGST          float64   `json:"sgst"`
	IGST          float64   `json:"igst"`
	GST           float64   `json:"gst"`
	Net           float64   `json:"net"`
}

// GetGST GET /api/v1/reports/gst?from=2026-04-01&to=2027-03-31
func (h *Handler) GetGST(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}
	from, to, ok := reportRange(c)
	if !ok {
		return
	}
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT i.id, i.invoice_number, COALESCE(NULLIF(c.company_name, ''), c.name, 'Customer'),
			i.issued_date, i.is_same_state,
			i.subtotal::float8, i.cgst_amount::float8, i.sgst_amount::float8, i.igst_amount::float8, i.tax_amount::float8,
			COALESCE(NULLIF(i.grand_total, 0), i.total)::float8
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		WHERE i.organisation_id = $1
		  AND i.issued_date BETWEEN $2 AND $3
		  AND i.status <> 'cancelled'
		ORDER BY i.issued_date, i.invoice_number
	`, orgID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load the GST report")
		return
	}
	defer rows.Close()
	lines := []gstLine{}
	for rows.Next() {
		var line gstLine
		var issued time.Time
		if err := rows.Scan(&line.ID, &line.InvoiceNumber, &line.CustomerName, &issued, &line.SameState, &line.Taxable, &line.CGST, &line.SGST, &line.IGST, &line.GST, &line.Net); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read the GST report")
			return
		}
		line.IssuedDate = issued.Format("2006-01-02")
		lines = append(lines, line)
	}
	response.Success(c, http.StatusOK, "GST report", lines)
}

func reportRange(c *gin.Context) (time.Time, time.Time, bool) {
	fromRaw := c.Query("from")
	toRaw := c.Query("to")
	if fromRaw == "" || toRaw == "" {
		response.Error(c, http.StatusBadRequest, "from and to dates are required")
		return time.Time{}, time.Time{}, false
	}
	from, err := time.Parse("2006-01-02", fromRaw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid from date")
		return time.Time{}, time.Time{}, false
	}
	to, err := time.Parse("2006-01-02", toRaw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid to date")
		return time.Time{}, time.Time{}, false
	}
	if to.Before(from) {
		response.Error(c, http.StatusBadRequest, "The end date is before the start date")
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}
