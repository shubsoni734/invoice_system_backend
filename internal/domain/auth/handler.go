package auth

import (
	"context"
	"net/http"
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	adminauthdb "github.com/your-org/invoice-backend/internal/domain/auth/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q          *adminauthdb.Queries
	jwtManager *utils.JWTManager
}

func NewHandler(q *adminauthdb.Queries, jwtManager *utils.JWTManager) *Handler {
	return &Handler{q: q, jwtManager: jwtManager}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := context.Background()

	user, err := h.q.GetUserByEmail(ctx, req.Email)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !user.IsActive {
		response.Error(c, http.StatusForbidden, "Account is inactive")
		return
	}
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		response.Error(c, http.StatusForbidden, "Account is locked. Try again later")
		return
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		_ = h.q.IncrementFailedAttempts(ctx, user.ID)
		response.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	_ = h.q.ResetFailedAttempts(ctx, user.ID)

	idStr := user.ID.String()
	orgStr := user.OrganisationID.String()

	accessToken, err := h.jwtManager.GenerateToken(&utils.Claims{
		UserID: idStr,
		OrgID:  orgStr,
		Role:   user.Role,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(&utils.Claims{
		UserID: idStr,
		OrgID:  orgStr,
		Role:   user.Role,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Parse client IP for storage
	ip, _ := netip.ParseAddr(c.ClientIP())
	uaStr := c.Request.UserAgent()
	var ua *string
	if uaStr != "" {
		ua = &uaStr
	}

	var ipType *netip.Addr
	if ip.IsValid() {
		ipType = &ip
	}

	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true}

	_, err = h.q.CreateRefreshToken(ctx, adminauthdb.CreateRefreshTokenParams{
		UserID:       user.ID,
		TokenHash:    utils.HashToken(refreshToken),
		ExpiresAt:    expiresAt,
		IpAddress:    ipType,
		UserAgent:    ua,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":              user.ID,
			"organisation_id": user.OrganisationID,
			"email":           user.Email,
			"name":            user.Name,
			"role":            user.Role,
		},
	})
}

// GetMe GET /api/v1/auth/me
func (h *Handler) GetMe(c *gin.Context) {
	rawID := c.GetString(constants.CtxUserID)
	if rawID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	user, err := h.q.GetUserByID(context.Background(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}

	response.Success(c, http.StatusOK, "User retrieved", gin.H{
		"id":              user.ID,
		"organisation_id": user.OrganisationID,
		"email":           user.Email,
		"name":            user.Name,
		"role":            user.Role,
		"is_active":       user.IsActive,
		"created_at":      user.CreatedAt.Time,
	})
}

// Logout POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	rawID := c.GetString(constants.CtxUserID)
	if rawID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	_ = h.q.RevokeAllUserTokens(context.Background(), id)
	response.Success(c, http.StatusOK, "Logout successful", nil)
}
