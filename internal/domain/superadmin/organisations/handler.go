package organisations

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	orgdb "github.com/your-org/invoice-backend/internal/domain/superadmin/organisations/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *orgdb.Queries
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{q: orgdb.New(db)}
}

type CreateOrgRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type ApplySubscriptionRequest struct {
	PlanID string `json:"plan_id"`
}

// CreateOrganisation POST /superadmin/organisations
func (h *Handler) CreateOrganisation(c *gin.Context) {
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	rawID, _ := c.Get(constants.CtxSuperAdminID)
	ctx := context.Background()

	slug := utils.SanitizeSlug(req.Name)
	slugExists, _ := h.q.OrgSlugExists(ctx, slug)
	if slugExists {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	// Convert superAdminID string to pgtype.UUID
	var createdBy pgtype.UUID
	if rawID != nil && rawID != "" {
		if id, err := uuid.Parse(rawID.(string)); err == nil {
			createdBy = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	row, err := h.q.CreateOrganisation(ctx, orgdb.CreateOrganisationParams{
		Name:                  req.Name,
		Slug:                  slug,
		Email:                 nullableStr(req.Email),
		Phone:                 nullableStr(req.Phone),
		Address:               nullableStr(req.Address),
		CreatedBySuperAdminID: createdBy,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create organisation: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Organisation created successfully", gin.H{
		"id":    row.ID,
		"name":  row.Name,
		"slug":  row.Slug,
		"email": row.Email,
	})
}

// ListOrganisations GET /superadmin/organisations
func (h *Handler) ListOrganisations(c *gin.Context) {
	pagination := utils.GetPaginationParams(c)
	ctx := context.Background()

	total, _ := h.q.CountOrganisations(ctx)

	orgs, err := h.q.ListOrganisations(ctx, orgdb.ListOrganisationsParams{
		Limit:  int32(pagination.PerPage),
		Offset: int32(pagination.Offset),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch organisations")
		return
	}

	// Ensure empty slice instead of nil for JSON
	if orgs == nil {
		orgs = []orgdb.ListOrganisationsRow{}
	}

	response.Success(c, http.StatusOK, "Organisations retrieved", gin.H{
		"organisations": orgs,
		"pagination": gin.H{
			"page":        pagination.Page,
			"per_page":    pagination.PerPage,
			"total":       total,
			"total_pages": utils.CalculateTotalPages(int64(total), pagination.PerPage),
		},
	})
}

// GetOrganisation GET /superadmin/organisations/:id
func (h *Handler) GetOrganisation(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid organisation ID")
		return
	}

	ctx := context.Background()

	org, err := h.q.GetOrganisationByID(ctx, orgID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Organisation not found")
		return
	}

	var sub interface{}
	subRow, subErr := h.q.GetActiveSubscription(ctx, orgID)
	if subErr == nil {
		sub = gin.H{
			"plan":       subRow.PlanName,
			"status":     subRow.Status,
			"period_end": subRow.CurrentPeriodEnd.Time,
		}
	}

	response.Success(c, http.StatusOK, "Organisation retrieved", gin.H{
		"organisation": org,
		"subscription": sub,
	})
}

// ApplySubscription POST /superadmin/organisations/:id/subscription
func (h *Handler) ApplySubscription(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid organisation ID")
		return
	}

	var req ApplySubscriptionRequest
	_ = c.ShouldBindJSON(&req)

	ctx := context.Background()

	exists, _ := h.q.OrgExists(ctx, orgID)
	if !exists {
		response.Error(c, http.StatusNotFound, "Organisation not found")
		return
	}

	var planID uuid.UUID

	if req.PlanID != "" {
		pid, parseErr := uuid.Parse(req.PlanID)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		planID, err = h.q.GetPlanByID(ctx, pid)
		if err != nil {
			response.Error(c, http.StatusNotFound, "Plan not found or inactive")
			return
		}
	} else {
		// Get or create the default Unlimited plan
		planID, err = h.q.GetUnlimitedPlan(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				planID, err = h.q.CreateUnlimitedPlan(ctx)
				if err != nil {
					response.Error(c, http.StatusInternalServerError, "Failed to create default plan: "+err.Error())
					return
				}
			} else {
				response.Error(c, http.StatusInternalServerError, "Database error")
				return
			}
		}
	}

	// Cancel any existing active subscription
	_ = h.q.CancelActiveSubscriptions(ctx, orgID)

	subID, err := h.q.CreateSubscription(ctx, orgdb.CreateSubscriptionParams{
		OrganisationID: orgID,
		PlanID:         planID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to apply subscription: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Subscription applied successfully", gin.H{
		"subscription_id": subID,
		"organisation_id": orgID,
		"plan_id":         planID,
		"status":          "active",
		"period_end":      "unlimited",
	})
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
