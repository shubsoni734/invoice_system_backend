package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	adminauthdb "github.com/your-org/invoice-backend/internal/domain/auth/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
)

func RegisterRoutes(
	public *gin.RouterGroup,
	protected *gin.RouterGroup,
	db *pgxpool.Pool,
	jwtManager *utils.JWTManager,
) {
	q := adminauthdb.New(db)
	handler := NewHandler(q, jwtManager)

	// Public routes
	public.POST("/login", handler.Login)
	public.POST("/forgot-password", handler.ForgotPassword)
	public.POST("/reset-password", handler.ResetPassword)

	// Protected routes
	protected.GET("/auth/me", handler.GetMe)
	protected.POST("/auth/logout", handler.Logout)
}
