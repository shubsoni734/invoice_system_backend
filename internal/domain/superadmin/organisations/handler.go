package organisations

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	orgdb "github.com/your-org/invoice-backend/internal/domain/superadmin/organisations/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
)

type Handler struct {
	q *orgdb.Queries
}

func NewHandler(q *orgdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateOrgRequest struct {
	Name    string             `json:"name" binding:"required"`
	Email   string             `json:"email" binding:"required,email"`
	Phone   *string            `json:"phone"`
	Address *string            `json:"address"`
	LogoUrl *string            `json:"logo_url"`
}

// CreateOrganisation POST /superadmin/organisations
func (h *Handler) CreateOrganisation(c *gin.Context) {
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	// Generate Slug
	slug := req.Name // For simplicity, we use the name as the slug for now.
	// In a real app, logic to generate a unique slug would go here.

	row, err := h.q.CreateOrganisation(ctx, orgdb.CreateOrganisationParams{
		Name:                  req.Name,
		Slug:                  slug,
		Email:                 &req.Email,
		Phone:                 req.Phone,
		Address:               req.Address,
		CreatedBySuperAdminID: pgtype.UUID{Valid: false}, // Placeholder for Creator
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	// Create Default Subscription (Free Plan)
	planID, err := h.q.GetUnlimitedPlan(ctx) // In this example we use an "Unlimited" plan
	if err != nil {
		// If no plan, we create it just for testing
		planID, _ = h.q.CreateUnlimitedPlan(ctx)
	}

	_, _ = h.q.CreateSubscription(ctx, orgdb.CreateSubscriptionParams{
		OrganisationID: row.ID,
		PlanID:         planID,
	})

	response.Success(c, http.StatusCreated, "Organisation created successfully", row)
}

// ListOrganisations GET /superadmin/organisations
func (h *Handler) ListOrganisations(c *gin.Context) {
	// Simple pagination
	orgs, err := h.q.ListOrganisations(c, orgdb.ListOrganisationsParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list organisations")
		return
	}

	response.Success(c, http.StatusOK, "Organisations retrieved", orgs)
}

// GetOrganisationByID GET /superadmin/organisations/:id
func (h *Handler) GetOrganisationByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	org, err := h.q.GetOrganisationByID(c, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Organisation not found")
		return
	}

	// Fetch active subscription details
	subscription, _ := h.q.GetActiveSubscription(c, id)

	response.Success(c, http.StatusOK, "Organisation retrieved", gin.H{
		"organisation": org,
		"subscription": subscription,
	})
}

type ApplySubscriptionRequest struct {
	PlanID *uuid.UUID `json:"plan_id"`
}

// ApplySubscription POST /superadmin/organisations/:id/subscription
func (h *Handler) ApplySubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req ApplySubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// 1. Verify org exists
	exists, err := h.q.OrgExists(ctx, id)
	if err != nil || !exists {
		response.Error(c, http.StatusNotFound, "Organisation not found")
		return
	}

	// 2. Get Plan ID
	var planID uuid.UUID
	if req.PlanID != nil {
		planID = *req.PlanID
		// Verify plan exists
		_, err = h.q.GetPlanByID(ctx, planID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid plan ID")
			return
		}
	} else {
		// Use default Unlimited plan
		planID, err = h.q.GetUnlimitedPlan(ctx)
		if err != nil {
			planID, _ = h.q.CreateUnlimitedPlan(ctx)
		}
	}

	// 3. Cancel existing active subscriptions
	_ = h.q.CancelActiveSubscriptions(ctx, id)

	// 4. Create new subscription
	subID, err := h.q.CreateSubscription(ctx, orgdb.CreateSubscriptionParams{
		OrganisationID: id,
		PlanID:         planID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to apply subscription")
		return
	}

	response.Success(c, http.StatusOK, "Subscription applied successfully", gin.H{
		"subscription_id": subID,
	})
}
