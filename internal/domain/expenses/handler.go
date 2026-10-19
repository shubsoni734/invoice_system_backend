package expenses

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/domain/notify"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

var categories = map[string]struct{}{
	"office_supplies": {}, "travel": {}, "utilities": {}, "rent": {}, "software": {},
	"marketing": {}, "salary": {}, "professional_fees": {}, "maintenance": {}, "miscellaneous": {},
}

var modes = map[string]struct{}{
	"cash": {}, "upi": {}, "bank_transfer": {}, "online": {}, "cheque": {},
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type expenseRequest struct {
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	VendorName  string  `json:"vendor_name"`
	GSTNumber   string  `json:"gst_number"`
	Amount      float64 `json:"amount"`
	GSTAmount   float64 `json:"gst_amount"`
	PaymentMode string  `json:"payment_mode"`
	PaymentDate string  `json:"payment_date"`
	Notes       string  `json:"notes"`
}

type expenseRow struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	VendorName  string    `json:"vendor_name"`
	GSTNumber   string    `json:"gst_number"`
	Amount      float64   `json:"amount"`
	GSTAmount   float64   `json:"gst_amount"`
	TotalAmount float64   `json:"total_amount"`
	PaymentMode string    `json:"payment_mode"`
	PaymentDate string    `json:"payment_date"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

const expenseColumns = `
	id, title, category, vendor_name, COALESCE(gst_number, ''),
	amount::float8, gst_amount::float8, total_amount::float8,
	payment_mode, payment_date, COALESCE(notes, ''), created_at`

func (h *Handler) List(c *gin.Context) {
	orgID, ok := orgID(c)
	if !ok {
		return
	}
	rows, err := h.db.Query(c.Request.Context(), `SELECT `+expenseColumns+` FROM expenses WHERE organisation_id = $1 ORDER BY payment_date DESC, created_at DESC`, orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load expenses")
		return
	}
	defer rows.Close()
	list := []expenseRow{}
	for rows.Next() {
		item, err := scanExpense(rows)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read expenses")
			return
		}
		list = append(list, item)
	}
	response.Success(c, http.StatusOK, "Expenses retrieved", list)
}

func (h *Handler) Create(c *gin.Context) {
	orgID, ok := orgID(c)
	if !ok {
		return
	}
	req, payDate, ok := bindExpense(c)
	if !ok {
		return
	}
	userID, _ := uuid.Parse(c.GetString(constants.CtxUserID))
	item, err := scanExpense(h.db.QueryRow(c.Request.Context(), `
		INSERT INTO expenses (
			organisation_id, title, category, vendor_name, gst_number,
			amount, gst_amount, total_amount, payment_mode, payment_date, notes, created_by
		) VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9,$10,NULLIF($11,''),$12)
		RETURNING `+expenseColumns,
		orgID, req.Title, req.Category, req.VendorName, strings.TrimSpace(req.GSTNumber),
		req.Amount, req.GSTAmount, req.Amount+req.GSTAmount, req.PaymentMode, payDate, strings.TrimSpace(req.Notes),
		nullableUser(userID),
	))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save expense")
		return
	}
	notify.Publish(orgID, "Expense added", req.Title+" was recorded.", "info")
	response.Success(c, http.StatusCreated, "Expense added", item)
}

func (h *Handler) Update(c *gin.Context) {
	orgID, ok := orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}
	req, payDate, ok := bindExpense(c)
	if !ok {
		return
	}
	item, err := scanExpense(h.db.QueryRow(c.Request.Context(), `
		UPDATE expenses SET
			title = $3, category = $4, vendor_name = $5, gst_number = NULLIF($6, ''),
			amount = $7, gst_amount = $8, total_amount = $9, payment_mode = $10,
			payment_date = $11, notes = NULLIF($12, ''), updated_at = NOW()
		WHERE id = $1 AND organisation_id = $2
		RETURNING `+expenseColumns,
		id, orgID, req.Title, req.Category, req.VendorName, strings.TrimSpace(req.GSTNumber),
		req.Amount, req.GSTAmount, req.Amount+req.GSTAmount, req.PaymentMode, payDate, strings.TrimSpace(req.Notes),
	))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Expense not found")
		return
	}
	notify.Publish(orgID, "Expense updated", req.Title+" was updated.", "info")
	response.Success(c, http.StatusOK, "Expense updated", item)
}

func (h *Handler) Delete(c *gin.Context) {
	orgID, ok := orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}
	tag, err := h.db.Exec(c.Request.Context(), `DELETE FROM expenses WHERE id = $1 AND organisation_id = $2`, id, orgID)
	if err != nil || tag.RowsAffected() == 0 {
		response.Error(c, http.StatusNotFound, "Expense not found")
		return
	}
	notify.Publish(orgID, "Expense removed", "An expense was removed.", "warning")
	response.Success(c, http.StatusOK, "Expense deleted", nil)
}

func orgID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(constants.CtxOrgID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return uuid.Nil, false
	}
	return id, true
}

func bindExpense(c *gin.Context) (expenseRequest, pgtype.Date, bool) {
	var req expenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return req, pgtype.Date{}, false
	}
	req.Title = strings.TrimSpace(req.Title)
	req.VendorName = strings.TrimSpace(req.VendorName)
	req.Category = strings.TrimSpace(req.Category)
	req.PaymentMode = strings.TrimSpace(req.PaymentMode)
	if req.Title == "" || req.VendorName == "" {
		response.Error(c, http.StatusBadRequest, "Title and vendor are required")
		return req, pgtype.Date{}, false
	}
	if _, ok := categories[req.Category]; !ok {
		response.Error(c, http.StatusBadRequest, "Choose a valid category")
		return req, pgtype.Date{}, false
	}
	if _, ok := modes[req.PaymentMode]; !ok {
		response.Error(c, http.StatusBadRequest, "Choose a valid payment mode")
		return req, pgtype.Date{}, false
	}
	if req.Amount < 0 || req.GSTAmount < 0 {
		response.Error(c, http.StatusBadRequest, "Amounts cannot be negative")
		return req, pgtype.Date{}, false
	}
	parsed, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment date")
		return req, pgtype.Date{}, false
	}
	return req, pgtype.Date{Time: parsed, Valid: true}, true
}

func nullableUser(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExpense(row scanner) (expenseRow, error) {
	var item expenseRow
	var payDate time.Time
	err := row.Scan(
		&item.ID, &item.Title, &item.Category, &item.VendorName, &item.GSTNumber,
		&item.Amount, &item.GSTAmount, &item.TotalAmount,
		&item.PaymentMode, &payDate, &item.Notes, &item.CreatedAt,
	)
	if err != nil {
		return item, err
	}
	item.PaymentDate = payDate.Format("2006-01-02")
	return item, nil
}

