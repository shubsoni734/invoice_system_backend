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
	whatsappdb "github.com/your-org/invoice-backend/internal/domain/whatsapp/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q          *whatsappdb.Queries
	db         *pgxpool.Pool
	apiURL     string
	apiKey     string
}

func NewHandler(q *whatsappdb.Queries, db *pgxpool.Pool, apiURL, apiKey string) *Handler {
	return &Handler{q: q, db: db, apiURL: apiURL, apiKey: apiKey}
}

type SendWhatsAppRequest struct {
	InvoiceID      uuid.UUID `json:"invoice_id" binding:"required"`
	RecipientPhone string    `json:"recipient_phone" binding:"required"`
	Message        string    `json:"message" binding:"required"`
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
	status := constants.WASent
	var errMsg *string
	sentAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	// Attempt to send via external WhatsApp API if configured
	if h.apiURL != "" && h.apiKey != "" {
		if sendErr := h.sendViaAPI(req.RecipientPhone, req.Message); sendErr != nil {
			status = constants.WAFailed
			msg := sendErr.Error()
			errMsg = &msg
			sentAt = pgtype.Timestamptz{}
		}
	}

	log, err := h.q.CreateWhatsappLog(ctx, whatsappdb.CreateWhatsappLogParams{
		OrganisationID: orgID,
		InvoiceID:      req.InvoiceID,
		RecipientPhone: req.RecipientPhone,
		Message:        req.Message,
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
