package settings

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/domain/notify"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q  *settingsdb.Queries
	db *pgxpool.Pool
}

func NewHandler(q *settingsdb.Queries, db *pgxpool.Pool) *Handler {
	return &Handler{q: q, db: db}
}

type SettingsRequest struct {
	BusinessName            *string    `json:"business_name"`
	BusinessEmail           *string    `json:"business_email"`
	BusinessPhone           *string    `json:"business_phone"`
	BusinessAddress         *string    `json:"business_address"`
	LogoUrl                 *string    `json:"logo_url"`
	Currency                string     `json:"currency"`
	DateFormat              string     `json:"date_format"`
	InvoicePrefix           string     `json:"invoice_prefix"`
	DefaultDueDays          int32      `json:"default_due_days"`
	DefaultTaxRate          float64    `json:"default_tax_rate"` // Usually mapped from numeric
	DefaultTemplateId       *uuid.UUID `json:"default_template_id"`
	WhatsappEnabled         bool       `json:"whatsapp_enabled"`
	WhatsappApiKey          *string    `json:"whatsapp_api_key"`
	WhatsappMessageTemplate string     `json:"whatsapp_message_template"`
}

// GetSettings GET /api/v1/settings
func (h *Handler) GetSettings(c *gin.Context) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDRaw)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	settingInfo, err := h.q.GetSettings(context.Background(), orgID)
	if err != nil {
		// If not found, a default one could be returned or just an empty one
		response.Success(c, http.StatusOK, "Settings not found, returning defaults", gin.H{})
		return
	}

	response.Success(c, http.StatusOK, "Settings retrieved successfully", settingInfo)
}

// UpsertSettings PUT /api/v1/settings
func (h *Handler) UpsertSettings(c *gin.Context) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDRaw)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// We can convert float64 to pgtype.Numeric manually
	var taxRate pgtype.Numeric
	taxRate.Scan(req.DefaultTaxRate)

	params := settingsdb.UpsertSettingsParams{
		OrganisationID:          orgID,
		BusinessName:            req.BusinessName,
		BusinessEmail:           req.BusinessEmail,
		BusinessPhone:           req.BusinessPhone,
		BusinessAddress:         req.BusinessAddress,
		LogoUrl:                 req.LogoUrl,
		Currency:                req.Currency,
		DateFormat:              req.DateFormat,
		InvoicePrefix:           req.InvoicePrefix,
		DefaultDueDays:          req.DefaultDueDays,
		DefaultTaxRate:          taxRate,
		WhatsappEnabled:         req.WhatsappEnabled,
		WhatsappApiKey:          req.WhatsappApiKey,
		WhatsappMessageTemplate: req.WhatsappMessageTemplate,
	}

	// Handle pointer to uuid
	// Based on emit_pointers_for_null_types, nullable UUID might be *uuid.UUID or uuid.NullUUID.
	// We need to convert from *uuid.UUID to pgtype.UUID
	if req.DefaultTemplateId != nil {
		params.DefaultTemplateID = pgtype.UUID{Bytes: *req.DefaultTemplateId, Valid: true}
	} else {
		params.DefaultTemplateID = pgtype.UUID{Valid: false}
	}

	settingInfo, err := h.q.UpsertSettings(context.Background(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	response.Success(c, http.StatusOK, "Settings updated successfully", settingInfo)
}

type channelSettings struct {
	NotifyEmailEnabled bool   `json:"notify_email_enabled"`
	PushEnabled        bool   `json:"push_enabled"`
	PushConfigured     bool   `json:"push_configured"`
	PushPublicKey      string `json:"push_public_key"`
	BrevoSenderEmail   string `json:"brevo_sender_email"`
	BrevoSenderName    string `json:"brevo_sender_name"`
	BrevoSMTPHost      string `json:"brevo_smtp_host"`
	BrevoSMTPPort      string `json:"brevo_smtp_port"`
	BrevoSMTPLogin     string `json:"brevo_smtp_login"`
	BrevoAPIKeySet     bool   `json:"brevo_api_key_set"`
	BrevoSMTPKeySet    bool   `json:"brevo_smtp_key_set"`
	BrevoAPIKeyHint    string `json:"brevo_api_key_hint"`
	BrevoSMTPKeyHint   string `json:"brevo_smtp_key_hint"`
}

type channelUpdate struct {
	NotifyEmailEnabled *bool   `json:"notify_email_enabled"`
	PushEnabled        *bool   `json:"push_enabled"`
	BrevoSenderEmail   *string `json:"brevo_sender_email"`
	BrevoSenderName    *string `json:"brevo_sender_name"`
	BrevoSMTPHost      *string `json:"brevo_smtp_host"`
	BrevoSMTPPort      *string `json:"brevo_smtp_port"`
	BrevoSMTPLogin     *string `json:"brevo_smtp_login"`
	BrevoSMTPKey       *string `json:"brevo_smtp_key"`
	BrevoAPIKey        *string `json:"brevo_api_key"`
}

// GetChannels GET /api/v1/settings/channels
func (h *Handler) GetChannels(c *gin.Context) {
	orgID, ok := orgFrom(c)
	if !ok {
		return
	}
	current, err := h.loadChannels(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load channel settings")
		return
	}
	response.Success(c, http.StatusOK, "Channel settings", current)
}

// UpdateChannels PUT /api/v1/settings/channels
func (h *Handler) UpdateChannels(c *gin.Context) {
	orgID, ok := orgFrom(c)
	if !ok {
		return
	}
	var req channelUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	ctx := c.Request.Context()
	if _, err := h.db.Exec(ctx, `
		INSERT INTO settings (organisation_id) VALUES ($1)
		ON CONFLICT (organisation_id) DO NOTHING
	`, orgID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to prepare channel settings")
		return
	}
	current, err := h.loadChannels(ctx, orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load channel settings")
		return
	}
	emailEnabled := current.NotifyEmailEnabled
	if req.NotifyEmailEnabled != nil {
		emailEnabled = *req.NotifyEmailEnabled
	}
	pushEnabled := current.PushEnabled
	if req.PushEnabled != nil {
		pushEnabled = *req.PushEnabled
	}
	senderEmail := pickText(req.BrevoSenderEmail, current.BrevoSenderEmail)
	senderName := pickText(req.BrevoSenderName, current.BrevoSenderName)
	smtpHost := pickText(req.BrevoSMTPHost, firstNonEmpty(current.BrevoSMTPHost, "smtp-relay.brevo.com"))
	smtpPort := pickText(req.BrevoSMTPPort, firstNonEmpty(current.BrevoSMTPPort, "587"))
	smtpLogin := pickText(req.BrevoSMTPLogin, current.BrevoSMTPLogin)

	if _, err := h.db.Exec(ctx, `
		UPDATE settings SET
			notify_email_enabled = $2,
			push_enabled = $3,
			brevo_sender_email = NULLIF($4, ''),
			brevo_sender_name = NULLIF($5, ''),
			brevo_smtp_host = NULLIF($6, ''),
			brevo_smtp_port = NULLIF($7, ''),
			brevo_smtp_login = NULLIF($8, ''),
			brevo_smtp_key = CASE WHEN $9 = '' THEN brevo_smtp_key ELSE $9 END,
			brevo_api_key = CASE WHEN $10 = '' THEN brevo_api_key ELSE $10 END,
			updated_at = NOW()
		WHERE organisation_id = $1
	`, orgID, emailEnabled, pushEnabled, senderEmail, senderName, smtpHost, smtpPort, smtpLogin, secretValue(req.BrevoSMTPKey), secretValue(req.BrevoAPIKey)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save channel settings")
		return
	}
	saved, err := h.loadChannels(ctx, orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load channel settings")
		return
	}
	response.Success(c, http.StatusOK, "Channel settings saved", saved)
}

// TestChannelEmail POST /api/v1/settings/channels/test
func (h *Handler) TestChannelEmail(c *gin.Context) {
	orgID, ok := orgFrom(c)
	if !ok {
		return
	}
	userID, err := uuid.Parse(c.GetString(constants.CtxUserID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return
	}
	var address string
	if err := h.db.QueryRow(c.Request.Context(), `SELECT email FROM users WHERE id = $1 AND organisation_id = $2`, userID, orgID).Scan(&address); err != nil {
		response.Error(c, http.StatusBadRequest, "Could not find your email address")
		return
	}
	client := notify.MailClientForOrg(c)
	if !client.Configured() {
		response.Error(c, http.StatusBadRequest, "Brevo is not configured. Add your SMTP key or API key, or set the server Brevo keys.")
		return
	}
	if err := client.SendTestEmail(address); err != nil {
		response.Error(c, http.StatusBadGateway, "Brevo rejected the test email")
		return
	}
	response.Success(c, http.StatusOK, "Test email sent", gin.H{"email": address})
}

func (h *Handler) loadChannels(ctx context.Context, orgID uuid.UUID) (channelSettings, error) {
	var apiKey, smtpKey string
	current := channelSettings{
		NotifyEmailEnabled: true,
		BrevoSMTPHost:      "smtp-relay.brevo.com",
		BrevoSMTPPort:      "587",
		PushConfigured:     notify.PushConfigured(),
		PushPublicKey:      notify.PublicKey(),
	}
	err := h.db.QueryRow(ctx, `
		SELECT COALESCE(notify_email_enabled, true), COALESCE(push_enabled, false),
		       COALESCE(brevo_sender_email, ''), COALESCE(brevo_sender_name, ''),
		       COALESCE(brevo_smtp_host, 'smtp-relay.brevo.com'), COALESCE(brevo_smtp_port, '587'),
		       COALESCE(brevo_smtp_login, ''), COALESCE(brevo_api_key, ''), COALESCE(brevo_smtp_key, '')
		FROM settings WHERE organisation_id = $1
	`, orgID).Scan(
		&current.NotifyEmailEnabled, &current.PushEnabled,
		&current.BrevoSenderEmail, &current.BrevoSenderName,
		&current.BrevoSMTPHost, &current.BrevoSMTPPort, &current.BrevoSMTPLogin,
		&apiKey, &smtpKey,
	)
	if err != nil {
		return current, nil
	}
	current.BrevoAPIKeySet = strings.TrimSpace(apiKey) != ""
	current.BrevoSMTPKeySet = strings.TrimSpace(smtpKey) != ""
	current.BrevoAPIKeyHint = secretHint(apiKey)
	current.BrevoSMTPKeyHint = secretHint(smtpKey)
	current.PushConfigured = notify.PushConfigured()
	current.PushPublicKey = notify.PublicKey()
	return current, nil
}

func orgFrom(c *gin.Context) (uuid.UUID, bool) {
	orgID, err := uuid.Parse(c.GetString(constants.CtxOrgID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return uuid.Nil, false
	}
	return orgID, true
}

func pickText(next *string, current string) string {
	if next == nil {
		return current
	}
	return strings.TrimSpace(*next)
}

func firstNonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func secretValue(value *string) string {
	if value == nil {
		return ""
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" || strings.HasPrefix(trimmed, "••••") {
		return ""
	}
	return trimmed
}

func secretHint(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "••••"
	}
	return "••••" + value[len(value)-4:]
}
