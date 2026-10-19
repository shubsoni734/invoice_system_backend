package access

import (
	"errors"
	"net/http"

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

type replaceRequest struct {
	Access []Grant `json:"access" binding:"required"`
}

func (h *Handler) Catalog(c *gin.Context) {
	response.Success(c, http.StatusOK, "Access catalog", Pages)
}

func (h *Handler) ListForRole(c *gin.Context) {
	orgID, roleID, ok := roleParams(c)
	if !ok {
		return
	}
	record, err := ListForRole(c.Request.Context(), h.db, orgID, roleID)
	if errors.Is(err, ErrRoleNotFound) {
		response.Error(c, http.StatusNotFound, "Role not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load role access")
		return
	}
	response.Success(c, http.StatusOK, "Role access", record)
}

func (h *Handler) ReplaceForRole(c *gin.Context) {
	orgID, roleID, ok := roleParams(c)
	if !ok {
		return
	}
	var req replaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	record, err := ReplaceForRole(c.Request.Context(), h.db, orgID, roleID, req.Access)
	if errors.Is(err, ErrRoleNotFound) {
		response.Error(c, http.StatusNotFound, "Role not found")
		return
	}
	if errors.Is(err, ErrSystemRole) {
		response.Error(c, http.StatusForbidden, "System role access cannot be changed")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save role access")
		return
	}
	response.Success(c, http.StatusOK, "Role access saved", record)
}

func roleParams(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := uuid.Parse(c.GetString(constants.CtxOrgID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return uuid.Nil, uuid.Nil, false
	}
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID")
		return uuid.Nil, uuid.Nil, false
	}
	return orgID, roleID, true
}
