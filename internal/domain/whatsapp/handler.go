package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	whatsappdb "github.com/your-org/invoice-backend/internal/domain/whatsapp/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q         *whatsappdb.Queries
	db        *pgxpool.Pool
	apiURL    string
	apiKey    string
	invoiceQ  *invoicesdb.Queries
	customerQ *customersdb.Queries
	settingsQ *settingsdb.Queries
}

func NewHandler(q *whatsappdb.Queries, db *pgxpool.Pool, apiURL, apiKey string) *Handler {
	return &Handler{
		q:         q,
		db:        db,
		apiURL:    apiURL,
		apiKey:    apiKey,
		invoiceQ:  invoicesdb.New(db),
		customerQ: customersdb.New(db),
		settingsQ: settingsdb.New(db),
	}
}

type SendWhatsAppRequest struct {
	InvoiceID      uuid.UUID `json:"invoice_id" binding:"required"`
	RecipientPhone string    `json:"recipient_phone"`
	Message        string    `json:"message"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(constants.CtxOrgID))
}

// GetWhatsAppLogs GET /api/v1/whatsapp/logs
func (h *Handler) GetWhatsAppLogs(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	logs, err := h.q.GetWhatsappLogs(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch WhatsApp logs")
		return
	}
	if logs == nil {
		logs = []whatsappdb.WhatsappLog{}
	}

	response.Success(c, http.StatusOK, "WhatsApp logs retrieved", logs)
}

// SendWhatsApp POST /api/v1/whatsapp/send
func (h *Handler) SendWhatsApp(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req SendWhatsAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()

	inv, err := h.invoiceQ.GetInvoiceByID(ctx, invoicesdb.GetInvoiceByIDParams{
		ID:             req.InvoiceID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}

	customer, _ := h.customerQ.GetCustomerByID(ctx, customersdb.GetCustomerByIDParams{
		ID:             inv.CustomerID,
		OrganisationID: orgID,
	})

	settings, _ := h.settingsQ.GetSettings(ctx, orgID)

	// Resolve phone: request field → customer phone
	phone := req.RecipientPhone
	if phone == "" && customer.Phone != nil && *customer.Phone != "" {
		phone = *customer.Phone
	}
	if phone == "" {
		response.Error(c, http.StatusBadRequest, "No recipient phone — provide recipient_phone or add phone to customer")
		return
	}

	// Build message from template if not provided
	message := req.Message
	if message == "" {
		message = buildWhatsAppMessage(settings, inv, customer)
	}

	status := constants.WASent
	var errMsg *string
	sentAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	if h.apiURL != "" && h.apiKey != "" {
		if sendErr := h.sendViaAPI(phone, message); sendErr != nil {
			status = constants.WAFailed
			msg := sendErr.Error()
			errMsg = &msg
			sentAt = pgtype.Timestamptz{}
		}
	}

	log, err := h.q.CreateWhatsappLog(ctx, whatsappdb.CreateWhatsappLogParams{
		OrganisationID: orgID,
		InvoiceID:      req.InvoiceID,
		RecipientPhone: phone,
		Message:        message,
		Status:         status,
		ErrorMessage:   errMsg,
		SentAt:         sentAt,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to log WhatsApp message")
		return
	}

	response.Success(c, http.StatusCreated, "WhatsApp message sent", log)
}

func buildWhatsAppMessage(settings settingsdb.Setting, inv invoicesdb.Invoice, customer customersdb.Customer) string {
	tmpl := settings.WhatsappMessageTemplate
	if tmpl == "" {
		tmpl = "Dear {{customer_name}}, please find invoice {{invoice_number}} for {{currency}} {{grand_total}}. Due: {{due_date}}."
	}

	dueDate := ""
	if inv.DueDate.Valid {
		dueDate = inv.DueDate.Time.Format("02 Jan 2006")
	}

	bizName := ""
	if settings.BusinessName != nil {
		bizName = *settings.BusinessName
	}

	r := strings.NewReplacer(
		"{{customer_name}}", customer.Name,
		"{{invoice_number}}", inv.InvoiceNumber,
		"{{total}}", fmt.Sprintf("%.2f", inv.Total),
		"{{grand_total}}", fmt.Sprintf("%.2f", inv.GrandTotal),
		"{{amount}}", fmt.Sprintf("%.2f", inv.GrandTotal),
		"{{currency}}", inv.Currency,
		"{{due_date}}", dueDate,
		"{{business_name}}", bizName,
	)
	return r.Replace(tmpl)
}

func (h *Handler) sendViaAPI(phone, message string) error {
	payload := map[string]string{
		"phone":   strings.TrimPrefix(phone, "+"),
		"message": message,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, h.apiURL+"/send", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}
	return nil
}
