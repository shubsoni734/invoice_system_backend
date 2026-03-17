package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	adminAuth "github.com/your-org/invoice-backend/internal/domain/auth"
	customers "github.com/your-org/invoice-backend/internal/domain/customers"
	invoices "github.com/your-org/invoice-backend/internal/domain/invoices"
	invoicesessions "github.com/your-org/invoice-backend/internal/domain/invoicesessions"
	services "github.com/your-org/invoice-backend/internal/domain/services"
	settings "github.com/your-org/invoice-backend/internal/domain/settings"
	superadminAuth "github.com/your-org/invoice-backend/internal/domain/superadmin/auth"
	superadminOrgs "github.com/your-org/invoice-backend/internal/domain/superadmin/organisations"
	templates "github.com/your-org/invoice-backend/internal/domain/templates"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
)

func RegisterRoutes(
	router *gin.Engine,
	db *pgxpool.Pool,
	orgJWT *utils.JWTManager,
	superJWT *utils.JWTManager,
	authRateLimiter *middleware.RateLimiter,
	apiRateLimiter *middleware.RateLimiter,
	superAdminIPAllowlist []string,
) {
	// Health
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Server is running", nil)
	})
	router.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			response.Error(c, http.StatusServiceUnavailable, "Database not ready")
			return
		}
		response.Success(c, http.StatusOK, "Server is ready", nil)
	})

	// Org API v1
	apiV1 := router.Group("/api/v1")
	{
		authPublic := apiV1.Group("/auth")
		authPublic.Use(middleware.RateLimit(authRateLimiter))

		protected := apiV1.Group("")
		protected.Use(middleware.RateLimit(apiRateLimiter))
		protected.Use(middleware.Auth(orgJWT))
		protected.Use(middleware.Tenant(db))

		adminAuth.RegisterRoutes(authPublic, protected, db, orgJWT)
		settings.RegisterRoutes(protected, db)
		customers.RegisterRoutes(protected, db)
		services.RegisterRoutes(protected, db)
		invoicesessions.RegisterRoutes(protected, db)
		templates.RegisterRoutes(protected, db)
		invoices.RegisterRoutes(protected, db)
	}

	// SuperAdmin
	superAdmin := router.Group("/superadmin")
	superAdmin.Use(middleware.RateLimit(authRateLimiter))
	{
		// Public: create + login
		superadminAuth.RegisterRoutes(superAdmin, db, superJWT, authRateLimiter)

		// Protected: requires valid superadmin JWT
		protected := superAdmin.Group("")
		protected.Use(middleware.SuperAuth(superJWT, superAdminIPAllowlist))
		{
			superadminOrgs.RegisterRoutes(protected.Group("/organisations"), db)
		}
	}
}
