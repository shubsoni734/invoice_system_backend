package settings

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *settingsdb.Queries
}

func NewHandler(q *settingsdb.Queries) *Handler {
	return &Handler{q: q}
}

type SettingsRequest struct {
	BusinessName            *string        `json:"business_name"`
	BusinessEmail           *string        `json:"business_email"`
	BusinessPhone           *string        `json:"business_phone"`
	BusinessAddress         *string        `json:"business_address"`
	LogoUrl                 *string        `json:"logo_url"`
	Currency                string         `json:"currency"`
	DateFormat              string         `json:"date_format"`
	InvoicePrefix           string         `json:"invoice_prefix"`
	DefaultDueDays          int32          `json:"default_due_days"`
	DefaultTaxRate          float64        `json:"default_tax_rate"` // Usually mapped from numeric
	DefaultTemplateId       *uuid.UUID     `json:"default_template_id"`
	WhatsappEnabled         bool           `json:"whatsapp_enabled"`
	WhatsappApiKey          *string        `json:"whatsapp_api_key"`
	WhatsappMessageTemplate string         `json:"whatsapp_message_template"`
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
