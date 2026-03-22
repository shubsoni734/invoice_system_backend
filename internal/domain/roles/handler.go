package roles

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	rolesdb "github.com/your-org/invoice-backend/internal/domain/roles/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *rolesdb.Queries
}

func NewHandler(q *rolesdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// ListRoles GET /api/v1/roles
func (h *Handler) ListRoles(c *gin.Context) {
	orgIDStr := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return
	}

	roles, err := h.q.ListRoles(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}
	if roles == nil {
		roles = []rolesdb.Role{}
	}

	response.Success(c, http.StatusOK, "Roles retrieved", roles)
}

// CreateRole POST /api/v1/roles
func (h *Handler) CreateRole(c *gin.Context) {
	orgIDStr := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return
	}

	// Permission check (Admin only)
	roleName := c.GetString(constants.CtxUserRole)
	if roleName != constants.RoleOrgAdmin {
		response.Error(c, http.StatusForbidden, "Only administrators can create roles")
		return
	}

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Check if role name already exists in this org
	_, err = h.q.GetRoleByName(context.Background(), rolesdb.GetRoleByNameParams{
		OrganisationID: orgID,
		Name:           req.Name,
	})
	if err == nil {
		response.Error(c, http.StatusConflict, "Role with this name already exists")
		return
	}

	role, err := h.q.CreateRole(context.Background(), rolesdb.CreateRoleParams{
		OrganisationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		IsSystem:       false,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create role")
		return
	}

	response.Success(c, http.StatusCreated, "Role created successfully", role)
}

// UpdateRole PUT /api/v1/roles/:id
func (h *Handler) UpdateRole(c *gin.Context) {
	orgIDStr := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return
	}

	// Permission check (Admin only)
	roleName := c.GetString(constants.CtxUserRole)
	if roleName != constants.RoleOrgAdmin {
		response.Error(c, http.StatusForbidden, "Only administrators can update roles")
		return
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Verify current role exists and isn't system-locked (if applicable)
	current, err := h.q.GetRoleByID(context.Background(), rolesdb.GetRoleByIDParams{
		ID:             roleID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Role not found")
		return
	}
	if current.IsSystem {
		response.Error(c, http.StatusForbidden, "System roles cannot be modified")
		return
	}

	role, err := h.q.UpdateRole(context.Background(), rolesdb.UpdateRoleParams{
		ID:             roleID,
		OrganisationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update role")
		return
	}

	response.Success(c, http.StatusOK, "Role updated successfully", role)
}

// DeleteRole DELETE /api/v1/roles/:id
func (h *Handler) DeleteRole(c *gin.Context) {
	orgIDStr := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return
	}

	// Permission check (Admin only)
	roleName := c.GetString(constants.CtxUserRole)
	if roleName != constants.RoleOrgAdmin {
		response.Error(c, http.StatusForbidden, "Only administrators can delete roles")
		return
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// SQL handles is_system check
	err = h.q.DeleteRole(context.Background(), rolesdb.DeleteRoleParams{
		ID:             roleID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	response.Success(c, http.StatusOK, "Role deleted successfully", nil)
}
// GetRole GET /api/v1/roles/:id
func (h *Handler) GetRole(c *gin.Context) {
	orgIDStr := c.GetString(constants.CtxOrgID)
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	role, err := h.q.GetRoleByID(context.Background(), rolesdb.GetRoleByIDParams{
		ID:             roleID,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	response.Success(c, http.StatusOK, "Role retrieved", role)
}
