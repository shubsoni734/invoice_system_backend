package dashboard

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type statusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type seriesPoint struct {
	Label   string  `json:"label"`
	Revenue float64 `json:"revenue"`
	Paid    float64 `json:"paid"`
}

type recentInvoice struct {
	ID            uuid.UUID `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	IssuedDate    string    `json:"issued_date"`
	Status        string    `json:"status"`
	GrandTotal    float64   `json:"grand_total"`
}

type statsPayload struct {
	Period    string          `json:"period"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Customers int64           `json:"customers"`
	Invoices  int64           `json:"invoices"`
	Revenue   float64         `json:"revenue"`
	Paid      float64         `json:"paid"`
	Pending   float64         `json:"pending"`
	Overdue   int64           `json:"overdue"`
	Status    []statusCount   `json:"status"`
	Series    []seriesPoint   `json:"series"`
	Recent    []recentInvoice `json:"recent"`
}

// Stats GET /api/v1/dashboard/stats?period=today|week|month|year
func (h *Handler) Stats(c *gin.Context) {
	orgID, err := uuid.Parse(c.GetString(constants.CtxOrgID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return
	}

	period := c.DefaultQuery("period", "month")
	from, to, bucket := periodRange(period, time.Now())
	ctx := c.Request.Context()
	payload := statsPayload{
		Period: period,
		From:   from.Format("2006-01-02"),
		To:     to.Format("2006-01-02"),
		Status: []statusCount{},
		Series: []seriesPoint{},
		Recent: []recentInvoice{},
	}

	err = h.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM customers WHERE organisation_id = $1 AND created_at::date BETWEEN $2 AND $3),
			(SELECT COUNT(*) FROM invoices WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3 AND status <> 'cancelled'),
			(SELECT COALESCE(SUM(COALESCE(NULLIF(grand_total, 0), total)), 0)::float8
			   FROM invoices WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3 AND status <> 'cancelled'),
			(SELECT COALESCE(SUM(amount), 0)::float8
			   FROM payments WHERE organisation_id = $1 AND payment_date BETWEEN $2 AND $3),
			(SELECT COALESCE(SUM(GREATEST(COALESCE(NULLIF(i.grand_total, 0), i.total) - COALESCE(paid.amount, 0), 0)), 0)::float8
			   FROM invoices i
			   LEFT JOIN (SELECT invoice_id, SUM(amount) AS amount FROM payments GROUP BY invoice_id) paid ON paid.invoice_id = i.id
			   WHERE i.organisation_id = $1 AND i.issued_date BETWEEN $2 AND $3 AND i.status <> 'cancelled'),
			(SELECT COUNT(*) FROM invoices
			   WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3
			     AND status NOT IN ('paid', 'cancelled') AND due_date < CURRENT_DATE)
	`, orgID, from, to).Scan(
		&payload.Customers, &payload.Invoices, &payload.Revenue, &payload.Paid, &payload.Pending, &payload.Overdue,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load dashboard statistics")
		return
	}

	statusRows, err := h.db.Query(ctx, `
		SELECT
			CASE
				WHEN status = 'paid' THEN 'paid'
				WHEN status = 'cancelled' THEN 'cancelled'
				WHEN status = 'draft' AND due_date < CURRENT_DATE THEN 'overdue'
				WHEN status IN ('sent', 'overdue') AND due_date < CURRENT_DATE THEN 'overdue'
				WHEN status = 'draft' THEN 'draft'
				ELSE 'sent'
			END AS bucket,
			COUNT(*)
		FROM invoices
		WHERE organisation_id = $1 AND issued_date BETWEEN $2 AND $3
		GROUP BY 1
	`, orgID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load invoice status")
		return
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var item statusCount
		if err := statusRows.Scan(&item.Status, &item.Count); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read invoice status")
			return
		}
		payload.Status = append(payload.Status, item)
	}

	seriesSQL := `
		SELECT to_char(d::date, 'DD Mon'),
			COALESCE((SELECT SUM(COALESCE(NULLIF(i.grand_total, 0), i.total)) FROM invoices i
				WHERE i.organisation_id = $1 AND i.issued_date = d::date AND i.status <> 'cancelled'), 0)::float8,
			COALESCE((SELECT SUM(p.amount) FROM payments p
				WHERE p.organisation_id = $1 AND p.payment_date = d::date), 0)::float8
		FROM generate_series($2::date, $3::date, '1 day') d
		ORDER BY d`
	if bucket == "month" {
		seriesSQL = `
			SELECT to_char(d, 'Mon'),
				COALESCE((SELECT SUM(COALESCE(NULLIF(i.grand_total, 0), i.total)) FROM invoices i
					WHERE i.organisation_id = $1 AND date_trunc('month', i.issued_date) = d AND i.status <> 'cancelled'), 0)::float8,
				COALESCE((SELECT SUM(p.amount) FROM payments p
					WHERE p.organisation_id = $1 AND date_trunc('month', p.payment_date) = d), 0)::float8
			FROM generate_series(date_trunc('month', $2::date), date_trunc('month', $3::date), '1 month') d
			ORDER BY d`
	}
	seriesRows, err := h.db.Query(ctx, seriesSQL, orgID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load revenue series")
		return
	}
	defer seriesRows.Close()
	for seriesRows.Next() {
		var point seriesPoint
		if err := seriesRows.Scan(&point.Label, &point.Revenue, &point.Paid); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read revenue series")
			return
		}
		payload.Series = append(payload.Series, point)
	}

	recentRows, err := h.db.Query(ctx, `
		SELECT i.id, i.invoice_number, i.customer_id, COALESCE(c.name, 'Customer'), i.issued_date,
			CASE
				WHEN i.status NOT IN ('paid', 'cancelled') AND i.due_date < CURRENT_DATE THEN 'overdue'
				ELSE i.status
			END,
			COALESCE(NULLIF(i.grand_total, 0), i.total)::float8
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		WHERE i.organisation_id = $1 AND i.issued_date BETWEEN $2 AND $3
		ORDER BY i.issued_date DESC, i.created_at DESC
		LIMIT 8
	`, orgID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load recent invoices")
		return
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var item recentInvoice
		var issued time.Time
		if err := recentRows.Scan(&item.ID, &item.InvoiceNumber, &item.CustomerID, &item.CustomerName, &issued, &item.Status, &item.GrandTotal); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read recent invoices")
			return
		}
		item.IssuedDate = issued.Format("2006-01-02")
		payload.Recent = append(payload.Recent, item)
	}

	response.Success(c, http.StatusOK, "Dashboard statistics", payload)
}

func periodRange(period string, now time.Time) (time.Time, time.Time, string) {
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	switch period {
	case "today":
		return today, today, "day"
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return today.AddDate(0, 0, -(weekday - 1)), today, "day"
	case "year":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc), today, "month"
	default:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc), today, "day"
	}
}
