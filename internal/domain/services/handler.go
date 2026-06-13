package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	servicesdb "github.com/your-org/invoice-backend/internal/domain/services/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

var validTaxRates = map[float64]bool{
	0: true, 5: true, 12: true, 18: true, 28: true,
}

type Handler struct {
	q *servicesdb.Queries
}

func NewHandler(q *servicesdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateServiceRequest struct {
	Type                string   `json:"type"`
	Name                string   `json:"name"`
	Description         *string  `json:"description"`
	Unit                *string  `json:"unit"`
	HSNCode             *string  `json:"hsn_code"`
	SACCode             *string  `json:"sac_code"`
	TaxPreference       string   `json:"tax_preference"`
	IntraStateTaxRate   *float64 `json:"intra_state_tax_rate"`
	InterStateTaxRate   *float64 `json:"inter_state_tax_rate"`
	SellingPrice        float64  `json:"selling_price"`
	Account             *string  `json:"account"`
	IsActive            *bool    `json:"is_active"`
}

type UpdateServiceRequest struct {
	Type                *string  `json:"type"`
	Name                *string  `json:"name"`
	Description         *string  `json:"description"`
	Unit                *string  `json:"unit"`
	HSNCode             *string  `json:"hsn_code"`
	SACCode             *string  `json:"sac_code"`
	TaxPreference       *string  `json:"tax_preference"`
	IntraStateTaxRate   *float64 `json:"intra_state_tax_rate"`
	InterStateTaxRate   *float64 `json:"inter_state_tax_rate"`
	SellingPrice        *float64 `json:"selling_price"`
	Account             *string  `json:"account"`
	IsActive            *bool    `json:"is_active"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

func validateCreateService(req *CreateServiceRequest) string {
	if strings.TrimSpace(req.Name) == "" {
		return "name is required"
	}

	svcType := strings.TrimSpace(req.Type)
	if svcType == "" {
		svcType = "service"
	}
	if svcType != "goods" && svcType != "service" {
		return "type must be 'goods' or 'service'"
	}

	tp := strings.TrimSpace(req.TaxPreference)
	if tp == "" {
		tp = "taxable"
	}
	if tp != "taxable" && tp != "tax_exempt" {
		return "tax_preference must be 'taxable' or 'tax_exempt'"
	}

	if req.SellingPrice < 0 {
		return "selling_price must be >= 0"
	}

	if req.IntraStateTaxRate != nil && !validTaxRates[*req.IntraStateTaxRate] {
		return "intra_state_tax_rate must be one of: 0, 5, 12, 18, 28"
	}

	if req.InterStateTaxRate != nil && !validTaxRates[*req.InterStateTaxRate] {
		return "inter_state_tax_rate must be one of: 0, 5, 12, 18, 28"
	}

	return ""
}

func validateUpdateService(req *UpdateServiceRequest) string {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return "name cannot be empty"
	}

	if req.Type != nil && *req.Type != "goods" && *req.Type != "service" {
		return "type must be 'goods' or 'service'"
	}

	if req.TaxPreference != nil && *req.TaxPreference != "taxable" && *req.TaxPreference != "tax_exempt" {
		return "tax_preference must be 'taxable' or 'tax_exempt'"
	}

	if req.SellingPrice != nil && *req.SellingPrice < 0 {
		return "selling_price must be >= 0"
	}

	if req.IntraStateTaxRate != nil && !validTaxRates[*req.IntraStateTaxRate] {
		return "intra_state_tax_rate must be one of: 0, 5, 12, 18, 28"
	}

	if req.InterStateTaxRate != nil && !validTaxRates[*req.InterStateTaxRate] {
		return "inter_state_tax_rate must be one of: 0, 5, 12, 18, 28"
	}

	return ""
}

// GetServices GET /api/v1/services
func (h *Handler) GetServices(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	search := c.Query("search")
	p := utils.GetPaginationParams(c)

	total, err := h.q.CountServices(context.Background(), servicesdb.CountServicesParams{
		OrganisationID: orgID,
		Search:         search,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to count services")
		return
	}

	res, err := h.q.GetServices(context.Background(), servicesdb.GetServicesParams{
		OrganisationID: orgID,
		Search:         search,
		Offset:         int32(p.Offset),
		PerPage:        int32(p.PerPage),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch services")
		return
	}

	if res == nil {
		res = []servicesdb.Service{}
	}

	response.SuccessWithMeta(c, http.StatusOK, "Services retrieved successfully", res, &response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: utils.CalculateTotalPages(total, p.PerPage),
	})
}

// GetServiceByID GET /api/v1/services/:id
func (h *Handler) GetServiceByID(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid service ID")
		return
	}

	res, err := h.q.GetServiceByID(context.Background(), servicesdb.GetServiceByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Service not found")
		return
	}

	response.Success(c, http.StatusOK, "Service retrieved successfully", res)
}

// CreateService POST /api/v1/services
func (h *Handler) CreateService(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply defaults before validation
	if strings.TrimSpace(req.Type) == "" {
		req.Type = "service"
	}
	if strings.TrimSpace(req.TaxPreference) == "" {
		req.TaxPreference = "taxable"
	}

	if errMsg := validateCreateService(&req); errMsg != "" {
		response.Error(c, http.StatusBadRequest, errMsg)
		return
	}

	// Resolve defaults for DB
	intraRate := 18.0
	if req.IntraStateTaxRate != nil {
		intraRate = *req.IntraStateTaxRate
	}
	interRate := 18.0
	if req.InterStateTaxRate != nil {
		interRate = *req.InterStateTaxRate
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	res, err := h.q.CreateService(context.Background(), servicesdb.CreateServiceParams{
		OrganisationID:    orgID,
		Type:              req.Type,
		Name:              strings.TrimSpace(req.Name),
		Description:       req.Description,
		Unit:              req.Unit,
		HsnCode:           req.HSNCode,
		SacCode:           req.SACCode,
		TaxPreference:     req.TaxPreference,
		IntraStateTaxRate: intraRate,
		InterStateTaxRate: interRate,
		SellingPrice:      req.SellingPrice,
		Account:           req.Account,
		IsActive:          isActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create service")
		return
	}

	response.Success(c, http.StatusCreated, "Service created successfully", res)
}

// UpdateService PUT /api/v1/services/:id
func (h *Handler) UpdateService(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid service ID")
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errMsg := validateUpdateService(&req); errMsg != "" {
		response.Error(c, http.StatusBadRequest, errMsg)
		return
	}

	res, err := h.q.UpdateService(context.Background(), servicesdb.UpdateServiceParams{
		ID:                id,
		OrganisationID:    orgID,
		Type:              req.Type,
		Name:              req.Name,
		Description:       req.Description,
		Unit:              req.Unit,
		HsnCode:           req.HSNCode,
		SacCode:           req.SACCode,
		TaxPreference:     req.TaxPreference,
		IntraStateTaxRate: req.IntraStateTaxRate,
		InterStateTaxRate: req.InterStateTaxRate,
		SellingPrice:      req.SellingPrice,
		Account:           req.Account,
		IsActive:          req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update service")
		return
	}

	response.Success(c, http.StatusOK, "Service updated successfully", res)
}

// DeleteService DELETE /api/v1/services/:id
func (h *Handler) DeleteService(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid service ID")
		return
	}

	if err := h.q.DeleteService(context.Background(), servicesdb.DeleteServiceParams{
		ID:             id,
		OrganisationID: orgID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete service")
		return
	}

	response.Success(c, http.StatusOK, "Service deleted successfully", nil)
}
