package auth

import (
	"context"
	"net/http"
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	authdb "github.com/your-org/invoice-backend/internal/domain/superadmin/auth/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q          *authdb.Queries
	jwtManager *utils.JWTManager
}

func NewHandler(q *authdb.Queries, jwtManager *utils.JWTManager) *Handler {
	return &Handler{q: q, jwtManager: jwtManager}
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
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	exists, err := h.q.SuperAdminEmailExists(ctx, req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		response.Error(c, http.StatusConflict, "SuperAdmin with this email already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	row, err := h.q.CreateSuperAdmin(ctx, authdb.CreateSuperAdminParams{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         constants.RoleSuperAdmin,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create superadmin: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "SuperAdmin created successfully", gin.H{
		"id":    row.ID,
		"email": row.Email,
		"role":  row.Role,
	})
}

// Login POST /superadmin/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	sa, err := h.q.GetSuperAdminByEmail(ctx, req.Email)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !sa.IsActive {
		response.Error(c, http.StatusForbidden, "Account is inactive")
		return
	}
	if sa.LockedUntil.Valid && sa.LockedUntil.Time.After(time.Now()) {
		response.Error(c, http.StatusForbidden, "Account is locked. Try again later")
		return
	}

	if !utils.CheckPassword(sa.PasswordHash, req.Password) {
		_ = h.q.IncrementFailedAttempts(ctx, sa.ID)
		response.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	_ = h.q.ResetFailedAttempts(ctx, sa.ID)

	idStr := sa.ID.String()
	accessToken, err := h.jwtManager.GenerateToken(&utils.Claims{
		SuperAdminID: idStr,
		Role:         sa.Role,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(&utils.Claims{
		SuperAdminID: idStr,
		Role:         sa.Role,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Parse client IP for storage
	ip, _ := netip.ParseAddr(c.ClientIP())
	ua := c.Request.UserAgent()
	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true}

	err = h.q.CreateSuperRefreshToken(ctx, authdb.CreateSuperRefreshTokenParams{
		SuperAdminID: sa.ID,
		TokenHash:    utils.HashToken(refreshToken),
		ExpiresAt:    expiresAt,
		IpAddress:    &ip,
		UserAgent:    &ua,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    sa.ID,
			"email": sa.Email,
			"role":  sa.Role,
		},
	})
}

// GetMe GET /superadmin/auth/me
func (h *Handler) GetMe(c *gin.Context) {
	rawID, _ := c.Get(constants.CtxSuperAdminID)
	if rawID == nil || rawID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(rawID.(string))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid superadmin ID")
		return
	}

	sa, err := h.q.GetSuperAdminByID(context.Background(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "SuperAdmin not found")
		return
	}

	response.Success(c, http.StatusOK, "SuperAdmin retrieved", gin.H{
		"id":         sa.ID,
		"email":      sa.Email,
		"role":       sa.Role,
		"is_active":  sa.IsActive,
		"created_at": sa.CreatedAt.Time,
	})
}

// Logout POST /superadmin/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	rawID, _ := c.Get(constants.CtxSuperAdminID)
	if rawID == nil || rawID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(rawID.(string))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid superadmin ID")
		return
	}

	_ = h.q.RevokeAllSuperRefreshTokens(context.Background(), id)
	response.Success(c, http.StatusOK, "Logout successful", nil)
}
