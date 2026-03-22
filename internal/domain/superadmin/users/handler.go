package users

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	usersdb "github.com/your-org/invoice-backend/internal/domain/superadmin/users/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *usersdb.Queries
}

func NewHandler(q *usersdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateUserRequest struct {
	Email    string     `json:"email" binding:"required,email"`
	Password string     `json:"password" binding:"required,min=8"`
	Name     string     `json:"name" binding:"required"`
	Role     string     `json:"role"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type SetStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// ListUsers GET /superadmin/organisations/:id/users
func (h *Handler) ListUsers(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid organisation ID")
		return
	}

	users, err := h.q.ListUsersByOrg(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	if users == nil {
		users = []usersdb.ListUsersByOrgRow{}
	}

	response.Success(c, http.StatusOK, "Users retrieved", users)
}

// CreateUser POST /superadmin/organisations/:id/users
func (h *Handler) CreateUser(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid organisation ID")
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	exists, _ := h.q.UserEmailExistsInOrg(ctx, usersdb.UserEmailExistsInOrgParams{
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
	
	role := req.Role
	var roleID pgtype.UUID
	
	if userCount == 0 {
		// First user is always admin
		role = constants.RoleOrgAdmin
		// Check if Admin role exists, if not create it
		adminRole, err := h.q.GetRoleByName(ctx, usersdb.GetRoleByNameParams{
			OrganisationID: orgID,
			Name:           "Admin",
		})
		if err != nil {
			// Create Admin role
			desc := "Administrator with full access"
			adminRole, _ = h.q.CreateRole(ctx, usersdb.CreateRoleParams{
				OrganisationID: orgID,
				Name:           "Admin",
				Description:    &desc,
				IsSystem:       true,
			})
		}
		roleID = pgtype.UUID{Bytes: adminRole.ID, Valid: true}
	} else {
		if req.RoleID != nil {
			roleID = pgtype.UUID{Bytes: *req.RoleID, Valid: true}
			// Verify role exists
			_, err = h.q.GetRoleByID(ctx, usersdb.GetRoleByIDParams{
				ID:             *req.RoleID,
				OrganisationID: orgID,
			})
			if err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid Role ID for this organization")
				return
			}
		} else if role == constants.RoleOrgAdmin {
			// If type is admin, find or create admin role
			adminRole, err := h.q.GetRoleByName(ctx, usersdb.GetRoleByNameParams{
				OrganisationID: orgID,
				Name:           "Admin",
			})
			if err == nil {
				roleID = pgtype.UUID{Bytes: adminRole.ID, Valid: true}
			}
		}
	}

	if role == "" {
		role = constants.RoleOrgViewer
	}

	user, err := h.q.CreateOrgUser(ctx, usersdb.CreateOrgUserParams{
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

// SetUserStatus PUT /superadmin/users/:id/status
func (h *Handler) SetUserStatus(c *gin.Context) {
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

	user, err := h.q.SetUserStatus(context.Background(), usersdb.SetUserStatusParams{
		ID:       userID,
		IsActive: req.IsActive,
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
