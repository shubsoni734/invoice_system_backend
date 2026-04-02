package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	orgusersdb "github.com/your-org/invoice-backend/internal/domain/users/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *orgusersdb.Queries
}

func NewHandler(q *orgusersdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateUserRequest struct {
	Email    string     `json:"email" binding:"required,email"`
	Password string     `json:"password" binding:"required,min=8"`
	Name     string     `json:"name" binding:"required"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type UpdateUserRequest struct {
	Name     string     `json:"name"`
	RoleID   *uuid.UUID `json:"role_id"`
	Password *string    `json:"password"`
}

type SetStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

// ListUsers GET /api/v1/users
func (h *Handler) ListUsers(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	users, err := h.q.ListUsersByOrg(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	if users == nil {
		users = []orgusersdb.ListUsersByOrgRow{}
	}

	response.Success(c, http.StatusOK, "Users retrieved", users)
}

// CreateUser POST /api/v1/users
func (h *Handler) CreateUser(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	exists, _ := h.q.UserEmailExistsInOrg(ctx, orgusersdb.UserEmailExistsInOrgParams{
		Email:          req.Email,
		OrganisationID: orgID,
	})
	if exists {
		response.Error(c, http.StatusConflict, "User with this email already exists in this organisation")
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Logic for Role and First User
	userCount, _ := h.q.GetOrgUserCount(ctx, orgID)
	
	role := constants.RoleOrgViewer // default
	var roleID pgtype.UUID
	
	if userCount == 0 {
		// First user is always admin
		role = constants.RoleOrgAdmin
		// Check if Admin role exists, if not create it
		adminRole, err := h.q.GetRoleByName(ctx, orgusersdb.GetRoleByNameParams{
			OrganisationID: orgID,
			Name:           "Admin",
		})
		if err != nil {
			// Create Admin role
			adminRole, _ = h.q.CreateDefaultAdminRole(ctx, orgID)
		}
		roleID = pgtype.UUID{Bytes: adminRole.ID, Valid: true}
		role = adminRole.Name
	} else if req.RoleID != nil {
		roleID = pgtype.UUID{Bytes: *req.RoleID, Valid: true}
		// Verify role exists and get its name
		roleData, err := h.q.GetRoleByID(ctx, orgusersdb.GetRoleByIDParams{
			ID:             *req.RoleID,
			OrganisationID: orgID,
		})
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid Role ID for this organization")
			return
		}
		role = roleData.Name
	}

	user, err := h.q.CreateOrgUser(ctx, orgusersdb.CreateOrgUserParams{
		OrganisationID: orgID,
		Email:          req.Email,
		PasswordHash:   hash,
		Name:           req.Name,
		Role:           role,
		RoleID:         roleID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	response.Success(c, http.StatusCreated, "User created successfully", user)
}

// UpdateUser PUT /api/v1/users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := c.Request.Context()

	// Logic for Role update
	role := ""
	var roleID pgtype.UUID
	if req.RoleID != nil {
		roleID = pgtype.UUID{Bytes: *req.RoleID, Valid: true}
		// Fetch role name
		roleData, err := h.q.GetRoleByID(ctx, orgusersdb.GetRoleByIDParams{
			ID:             *req.RoleID,
			OrganisationID: orgID,
		})
		if err == nil {
			role = roleData.Name
		}
	}

	user, err := h.q.UpdateOrgUser(ctx, orgusersdb.UpdateOrgUserParams{
		ID:             userID,
		OrganisationID: orgID,
		Name:           req.Name,
		Role:           role,
		RoleID:         roleID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", user)
}

// SetUserStatus PUT /api/v1/users/:id/status
func (h *Handler) SetUserStatus(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req SetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.q.SetUserStatus(c.Request.Context(), orgusersdb.SetUserStatusParams{
		ID:             userID,
		IsActive:       req.IsActive,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update user status")
		return
	}

	msg := "User disabled"
	if req.IsActive {
		msg = "User enabled"
	}
	response.Success(c, http.StatusOK, msg, user)
}
