package templates

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	templatesdb "github.com/your-org/invoice-backend/internal/domain/templates/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *templatesdb.Queries
}

func NewHandler(q *templatesdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateTemplateRequest struct {
	Name         string  `json:"name" binding:"required"`
	HtmlContent  string  `json:"html_content" binding:"required"`
	IsDefault    *bool   `json:"is_default"`
	ThumbnailUrl *string `json:"thumbnail_url"`
}

type UpdateTemplateRequest struct {
	Name         *string `json:"name"`
	HtmlContent  *string `json:"html_content"`
	IsDefault    *bool   `json:"is_default"`
	ThumbnailUrl *string `json:"thumbnail_url"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDRaw := c.GetString(constants.CtxUserID)
	return uuid.Parse(userIDRaw)
}

// GetTemplates GET /api/v1/templates
func (h *Handler) GetTemplates(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	res, err := h.q.GetTemplates(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch templates")
		return
	}

	if res == nil {
		res = []templatesdb.Template{}
	}

	response.Success(c, http.StatusOK, "Templates retrieved successfully", res)
}

// GetTemplateByID GET /api/v1/templates/:id
func (h *Handler) GetTemplateByID(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid template ID")
		return
	}

	res, err := h.q.GetTemplateByID(context.Background(), templatesdb.GetTemplateByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Template not found")
		return
	}

	response.Success(c, http.StatusOK, "Template retrieved successfully", res)
}

// CreateTemplate POST /api/v1/templates
func (h *Handler) CreateTemplate(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	isDef := false
	if req.IsDefault != nil {
		isDef = *req.IsDefault
	}

	createdBy := pgtype.UUID{Bytes: userID, Valid: true}

	res, err := h.q.CreateTemplate(context.Background(), templatesdb.CreateTemplateParams{
		OrganisationID: orgID,
		Name:           req.Name,
		HtmlContent:    req.HtmlContent,
		IsDefault:      isDef,
		ThumbnailUrl:   req.ThumbnailUrl,
		CreatedBy:      createdBy,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create template")
		return
	}

	response.Success(c, http.StatusCreated, "Template created successfully", res)
}

// UpdateTemplate PUT /api/v1/templates/:id
func (h *Handler) UpdateTemplate(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid template ID")
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.q.UpdateTemplate(context.Background(), templatesdb.UpdateTemplateParams{
		ID:             id,
		OrganisationID: orgID,
		Name:           req.Name,
		HtmlContent:    req.HtmlContent,
		IsDefault:      req.IsDefault,
		ThumbnailUrl:   req.ThumbnailUrl,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update template")
		return
	}

	response.Success(c, http.StatusOK, "Template updated successfully", res)
}

// DeleteTemplate DELETE /api/v1/templates/:id
func (h *Handler) DeleteTemplate(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid template ID")
		return
	}

	err = h.q.DeleteTemplate(context.Background(), templatesdb.DeleteTemplateParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete template")
		return
	}

	response.Success(c, http.StatusOK, "Template deleted successfully", nil)
}
