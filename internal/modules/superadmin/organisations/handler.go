package organisations

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/constants"
	"github.com/your-org/invoice-backend/internal/utils"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type CreateOrgRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type ApplySubscriptionRequest struct {
	PlanID string `json:"plan_id"` // optional — if empty, uses default unlimited plan
}

// CreateOrganisation POST /superadmin/organisations
func (h *Handler) CreateOrganisation(c *gin.Context) {
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	superAdminID, _ := c.Get(constants.CtxSuperAdminID)

	slug := generateSlug(req.Name)
	ctx := context.Background()

	// Ensure slug is unique
	var slugExists bool
	_ = h.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM organisations WHERE slug = $1)", slug).Scan(&slugExists)
	if slugExists {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	var id uuid.UUID
	err := h.db.QueryRow(ctx, `
		INSERT INTO organisations (name, slug, email, phone, address, status, created_by_super_admin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, NOW(), NOW())
		RETURNING id
	`, req.Name, slug, nullableStr(req.Email), nullableStr(req.Phone), nullableStr(req.Address), superAdminID).Scan(&id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create organisation: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Organisation created successfully", gin.H{
		"id":    id,
		"name":  req.Name,
		"slug":  slug,
		"email": req.Email,
	})
}

// ListOrganisations GET /superadmin/organisations
func (h *Handler) ListOrganisations(c *gin.Context) {
	pagination := utils.GetPaginationParams(c)
	ctx := context.Background()

	var total int
	_ = h.db.QueryRow(ctx, "SELECT COUNT(*) FROM organisations").Scan(&total)

	rows, err := h.db.Query(ctx, `
		SELECT id, name, slug, email, phone, status, created_at
		FROM organisations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pagination.PerPage, pagination.Offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch organisations")
		return
	}
	defer rows.Close()

	type OrgRow struct {
		ID        uuid.UUID  `json:"id"`
		Name      string     `json:"name"`
		Slug      string     `json:"slug"`
		Email     *string    `json:"email"`
		Phone     *string    `json:"phone"`
		Status    string     `json:"status"`
		CreatedAt time.Time  `json:"created_at"`
	}

	orgs := []OrgRow{}
	for rows.Next() {
		var o OrgRow
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.Email, &o.Phone, &o.Status, &o.CreatedAt); err != nil {
			continue
		}
		orgs = append(orgs, o)
	}

	utils.SuccessResponse(c, http.StatusOK, "Organisations retrieved", gin.H{
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
	orgID := c.Param("id")
	ctx := context.Background()

	type OrgDetail struct {
		ID        uuid.UUID  `json:"id"`
		Name      string     `json:"name"`
		Slug      string     `json:"slug"`
		Email     *string    `json:"email"`
		Phone     *string    `json:"phone"`
		Address   *string    `json:"address"`
		Status    string     `json:"status"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
	}

	var o OrgDetail
	err := h.db.QueryRow(ctx, `
		SELECT id, name, slug, email, phone, address, status, created_at, updated_at
		FROM organisations WHERE id = $1
	`, orgID).Scan(&o.ID, &o.Name, &o.Slug, &o.Email, &o.Phone, &o.Address, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Organisation not found")
		return
	}

	// Get active subscription
	var sub *gin.H
	var planName string
	var subStatus string
	var periodEnd time.Time
	subErr := h.db.QueryRow(ctx, `
		SELECT p.name, os.status, os.current_period_end
		FROM organisation_subscriptions os
		JOIN plans p ON p.id = os.plan_id
		WHERE os.organisation_id = $1 AND os.status = 'active'
		ORDER BY os.created_at DESC LIMIT 1
	`, orgID).Scan(&planName, &subStatus, &periodEnd)
	if subErr == nil {
		s := gin.H{"plan": planName, "status": subStatus, "period_end": periodEnd}
		sub = &s
	}

	utils.SuccessResponse(c, http.StatusOK, "Organisation retrieved", gin.H{
		"organisation": o,
		"subscription": sub,
	})
}

// ApplySubscription POST /superadmin/organisations/:id/subscription
func (h *Handler) ApplySubscription(c *gin.Context) {
	orgID := c.Param("id")
	var req ApplySubscriptionRequest
	_ = c.ShouldBindJSON(&req) // optional body

	ctx := context.Background()

	// Verify org exists
	var orgExists bool
	_ = h.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM organisations WHERE id = $1)", orgID).Scan(&orgExists)
	if !orgExists {
		utils.ErrorResponse(c, http.StatusNotFound, "Organisation not found")
		return
	}

	var planID uuid.UUID

	if req.PlanID != "" {
		// Use provided plan
		err := h.db.QueryRow(ctx, "SELECT id FROM plans WHERE id = $1 AND is_active = true", req.PlanID).Scan(&planID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Plan not found or inactive")
			return
		}
	} else {
		// Get or create the default "Unlimited" plan
		err := h.db.QueryRow(ctx, "SELECT id FROM plans WHERE name = 'Unlimited' LIMIT 1").Scan(&planID)
		if err != nil {
			// Create the unlimited plan
			err = h.db.QueryRow(ctx, `
				INSERT INTO plans (name, price_monthly, price_yearly, max_users, max_customers,
					max_invoices_per_month, max_storage_mb, whatsapp_enabled, custom_templates,
					api_access, is_active, created_at, updated_at)
				VALUES ('Unlimited', 0, 0, 999999, 999999, 999999, 999999, true, true, true, true, NOW(), NOW())
				RETURNING id
			`).Scan(&planID)
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create default plan: "+err.Error())
				return
			}
		}
	}

	// Deactivate any existing active subscription
	_, _ = h.db.Exec(ctx, `
		UPDATE organisation_subscriptions
		SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
		WHERE organisation_id = $1 AND status = 'active'
	`, orgID)

	// Create new subscription — active, no expiry (100 years)
	var subID uuid.UUID
	err := h.db.QueryRow(ctx, `
		INSERT INTO organisation_subscriptions
			(organisation_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NOW() + INTERVAL '100 years', NOW(), NOW())
		RETURNING id
	`, orgID, planID).Scan(&subID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to apply subscription: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Subscription applied successfully", gin.H{
		"subscription_id":  subID,
		"organisation_id":  orgID,
		"plan_id":          planID,
		"status":           "active",
		"period_end":       "unlimited",
	})
}

// helpers
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return strings.Trim(result.String(), "-")
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
