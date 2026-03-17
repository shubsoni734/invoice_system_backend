package customers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *customersdb.Queries
}

func NewHandler(q *customersdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateCustomerRequest struct {
	Name      string  `json:"name" binding:"required"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Address   *string `json:"address"`
	TaxNumber *string `json:"tax_number"`
	IsActive  *bool   `json:"is_active"`
}

type UpdateCustomerRequest struct {
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Address   *string `json:"address"`
	TaxNumber *string `json:"tax_number"`
	IsActive  *bool   `json:"is_active"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

// GetCustomers GET /api/v1/customers
func (h *Handler) GetCustomers(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	res, err := h.q.GetCustomers(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch customers")
		return
	}

	if res == nil {
		res = []customersdb.Customer{}
	}

	response.Success(c, http.StatusOK, "Customers retrieved successfully", res)
}

// GetCustomerByID GET /api/v1/customers/:id
func (h *Handler) GetCustomerByID(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	res, err := h.q.GetCustomerByID(context.Background(), customersdb.GetCustomerByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Customer not found")
		return
	}

	response.Success(c, http.StatusOK, "Customer retrieved successfully", res)
}

// CreateCustomer POST /api/v1/customers
func (h *Handler) CreateCustomer(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	res, err := h.q.CreateCustomer(context.Background(), customersdb.CreateCustomerParams{
		OrganisationID: orgID,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		TaxNumber:      req.TaxNumber,
		IsActive:       isActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create customer")
		return
	}

	response.Success(c, http.StatusCreated, "Customer created successfully", res)
}

// UpdateCustomer PUT /api/v1/customers/:id
func (h *Handler) UpdateCustomer(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.q.UpdateCustomer(context.Background(), customersdb.UpdateCustomerParams{
		ID:             id,
		OrganisationID: orgID,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		TaxNumber:      req.TaxNumber,
		IsActive:       req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update customer")
		return
	}

	response.Success(c, http.StatusOK, "Customer updated successfully", res)
}

// DeleteCustomer DELETE /api/v1/customers/:id
func (h *Handler) DeleteCustomer(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	err = h.q.DeleteCustomer(context.Background(), customersdb.DeleteCustomerParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete customer")
		return
	}

	response.Success(c, http.StatusOK, "Customer deleted successfully", nil)
}
