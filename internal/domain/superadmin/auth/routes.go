package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	authdb "github.com/your-org/invoice-backend/internal/domain/superadmin/auth/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, jwtManager *utils.JWTManager, authRateLimiter *middleware.RateLimiter) {
	q := authdb.New(db)
	handler := NewHandler(q, jwtManager)

	// Public routes
	public := router.Group("/auth")
	public.Use(middleware.RateLimit(authRateLimiter))
	{
		public.POST("/create", handler.CreateSuperAdmin)
		public.POST("/login", handler.Login)
	}

	// Protected routes — SuperAuth applied at router level in app/routes.go
	protected := router.Group("/auth")
	{
		protected.GET("/me", handler.GetMe)
		protected.POST("/logout", handler.Logout)
	}
}
