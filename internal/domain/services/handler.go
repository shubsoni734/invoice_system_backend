package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	servicesdb "github.com/your-org/invoice-backend/internal/domain/services/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *servicesdb.Queries
}

func NewHandler(q *servicesdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateServiceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	UnitPrice   float64 `json:"unit_price" binding:"required"`
	TaxRate     float64 `json:"tax_rate"`
	Unit        string  `json:"unit"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateServiceRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	UnitPrice   *float64 `json:"unit_price"`
	TaxRate     *float64 `json:"tax_rate"`
	Unit        *string  `json:"unit"`
	IsActive    *bool    `json:"is_active"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

func floatToNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	// Use string representation for maximum compatibility with pgx/v5 Scan
	strValue := fmt.Sprintf("%v", f)
	err := num.Scan(strValue)
	if err != nil {
		fmt.Printf("floatToNumeric conversion error for %v: %v\n", f, err)
		num.Valid = false
	} else {
		num.Valid = true
	}
	return num
}

// GetServices GET /api/v1/services
func (h *Handler) GetServices(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	res, err := h.q.GetServices(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch services")
		return
	}

	if res == nil {
		res = []servicesdb.Service{}
	}

	response.Success(c, http.StatusOK, "Services retrieved successfully", res)
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

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	unit := "unit"
	if req.Unit != "" {
		unit = req.Unit
	}

	res, err := h.q.CreateService(context.Background(), servicesdb.CreateServiceParams{
		OrganisationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		UnitPrice:      floatToNumeric(req.UnitPrice),
		TaxRate:        floatToNumeric(req.TaxRate),
		Unit:           unit,
		IsActive:       isActive,
	})
	if err != nil {
		fmt.Printf("CreateService Error: %v\n", err)
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create service: %v", err))
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

	var unitPrice pgtype.Numeric
	if req.UnitPrice != nil {
		unitPrice = floatToNumeric(*req.UnitPrice)
	}

	var taxRate pgtype.Numeric
	if req.TaxRate != nil {
		taxRate = floatToNumeric(*req.TaxRate)
	}

	res, err := h.q.UpdateService(context.Background(), servicesdb.UpdateServiceParams{
		ID:             id,
		OrganisationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		UnitPrice:      unitPrice,
		TaxRate:        taxRate,
		Unit:           req.Unit,
		IsActive:       req.IsActive,
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

	err = h.q.DeleteService(context.Background(), servicesdb.DeleteServiceParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete service")
		return
	}

	response.Success(c, http.StatusOK, "Service deleted successfully", nil)
}
