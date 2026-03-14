package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/constants"
	"github.com/your-org/invoice-backend/internal/utils"
)

type Handler struct {
	db         *pgxpool.Pool
	jwtManager *utils.JWTManager
}

func NewHandler(db *pgxpool.Pool, jwtManager *utils.JWTManager) *Handler {
	return &Handler{db: db, jwtManager: jwtManager}
}

type CreateSuperAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateSuperAdmin POST /superadmin/auth/create
func (h *Handler) CreateSuperAdmin(c *gin.Context) {
	var req CreateSuperAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	var exists bool
	if err := h.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM super_admins WHERE email = $1)", req.Email).Scan(&exists); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		utils.ErrorResponse(c, http.StatusConflict, "SuperAdmin with this email already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	var id uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO super_admins (email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id
	`, req.Email, hashedPassword, constants.RoleSuperAdmin).Scan(&id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create superadmin: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "SuperAdmin created successfully", gin.H{
		"id":    id,
		"email": req.Email,
		"role":  constants.RoleSuperAdmin,
	})
}

// Login POST /superadmin/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	var id uuid.UUID
	var passwordHash, role string
	var isActive bool
	var failedAttempts int
	var lockedUntil *time.Time

	err := h.db.QueryRow(ctx, `
		SELECT id, password_hash, role, is_active, failed_attempts, locked_until
		FROM super_admins WHERE email = $1
	`, req.Email).Scan(&id, &passwordHash, &role, &isActive, &failedAttempts, &lockedUntil)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !isActive {
		utils.ErrorResponse(c, http.StatusForbidden, "Account is inactive")
		return
	}
	if lockedUntil != nil && lockedUntil.After(time.Now()) {
		utils.ErrorResponse(c, http.StatusForbidden, "Account is locked. Try again later")
		return
	}

	if !utils.CheckPassword(passwordHash, req.Password) {
		_, _ = h.db.Exec(ctx, `
			UPDATE super_admins
			SET failed_attempts = failed_attempts + 1,
				locked_until = CASE WHEN failed_attempts >= 4 THEN NOW() + INTERVAL '15 minutes' ELSE NULL END
			WHERE id = $1
		`, id)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	_, _ = h.db.Exec(ctx, `
		UPDATE super_admins SET failed_attempts = 0, locked_until = NULL, last_login_at = NOW()
		WHERE id = $1
	`, id)

	idStr := id.String()
	accessToken, err := h.jwtManager.GenerateToken(&utils.Claims{
		SuperAdminID: idStr,
		Role:         role,
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(&utils.Claims{
		SuperAdminID: idStr,
		Role:         role,
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	_, err = h.db.Exec(ctx, `
		INSERT INTO super_refresh_tokens (super_admin_id, token_hash, expires_at, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, id, utils.HashToken(refreshToken), time.Now().Add(7*24*time.Hour), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    id,
			"email": req.Email,
			"role":  role,
		},
	})
}

// GetMe GET /superadmin/auth/me
func (h *Handler) GetMe(c *gin.Context) {
	superAdminID, _ := c.Get(constants.CtxSuperAdminID)
	if superAdminID == nil || superAdminID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := context.Background()

	var id uuid.UUID
	var email, role string
	var isActive bool
	var createdAt time.Time

	err := h.db.QueryRow(ctx, `
		SELECT id, email, role, is_active, created_at FROM super_admins WHERE id = $1
	`, superAdminID).Scan(&id, &email, &role, &isActive, &createdAt)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "SuperAdmin not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "SuperAdmin retrieved", gin.H{
		"id":         id,
		"email":      email,
		"role":       role,
		"is_active":  isActive,
		"created_at": createdAt,
	})
}

// Logout POST /superadmin/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	superAdminID, _ := c.Get(constants.CtxSuperAdminID)
	if superAdminID == nil || superAdminID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := context.Background()
	// Revoke all refresh tokens for this superadmin
	_, _ = h.db.Exec(ctx, `
		UPDATE super_refresh_tokens SET revoked_at = NOW() WHERE super_admin_id = $1 AND revoked_at IS NULL
	`, superAdminID)

	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}
