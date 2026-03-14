package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/middleware"
	"github.com/your-org/invoice-backend/internal/utils"
)

// RegisterRoutes registers all superadmin auth routes
func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, jwtManager *utils.JWTManager, authRateLimiter *middleware.RateLimiter) {
	handler := NewHandler(db, jwtManager)

	// Public routes (no authentication required)
	public := router.Group("/auth")
	public.Use(middleware.RateLimit(authRateLimiter))
	{
		public.POST("/create", handler.CreateSuperAdmin) // Create superadmin
		public.POST("/login", handler.Login)             // Login
	}

	// Protected routes (authentication required)
	protected := router.Group("/auth")
	protected.Use(middleware.RateLimit(authRateLimiter))
	// Note: SuperAuth middleware should be applied at the router level
	{
		protected.GET("/me", handler.GetMe)       // Get current user
		protected.POST("/logout", handler.Logout) // Logout
	}
}
