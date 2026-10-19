package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/domain/access"
	adminAuth "github.com/your-org/invoice-backend/internal/domain/auth"
	common "github.com/your-org/invoice-backend/internal/domain/common"
	customers "github.com/your-org/invoice-backend/internal/domain/customers"
	"github.com/your-org/invoice-backend/internal/domain/expenses"
	"github.com/your-org/invoice-backend/internal/domain/dashboard"
	invoices "github.com/your-org/invoice-backend/internal/domain/invoices"
	invoicesessions "github.com/your-org/invoice-backend/internal/domain/invoicesessions"
	"github.com/your-org/invoice-backend/internal/domain/notify"
	payments "github.com/your-org/invoice-backend/internal/domain/payments"
	invoicepdf "github.com/your-org/invoice-backend/internal/domain/pdf"
	reports "github.com/your-org/invoice-backend/internal/domain/reports"
	roles "github.com/your-org/invoice-backend/internal/domain/roles"
	services "github.com/your-org/invoice-backend/internal/domain/services"
	settings "github.com/your-org/invoice-backend/internal/domain/settings"
	superadminAuth "github.com/your-org/invoice-backend/internal/domain/superadmin/auth"
	superadminOrgs "github.com/your-org/invoice-backend/internal/domain/superadmin/organisations"
	superadminUsers "github.com/your-org/invoice-backend/internal/domain/superadmin/users"
	templates "github.com/your-org/invoice-backend/internal/domain/templates"
	users "github.com/your-org/invoice-backend/internal/domain/users"
	whatsapp "github.com/your-org/invoice-backend/internal/domain/whatsapp"
	"github.com/your-org/invoice-backend/internal/pkg/email"
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
	whatsAppAPIURL string,
	whatsAppAPIKey string,
	emailClient *email.Client,
	frontendURL string,
	environment string,
	bootstrapSecret string,
	enableTestEmail bool,
	emailTestToken string,
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

		if enableTestEmail {
			common.RegisterRoutes(apiV1, emailClient)
		}
		emailCheck := apiV1.Group("")
		emailCheck.Use(middleware.RateLimit(middleware.NewRateLimiter(5)))
		common.RegisterSampleRoute(emailCheck, emailClient, emailTestToken)
		adminAuth.RegisterRoutes(authPublic, protected, db, orgJWT, emailClient, frontendURL)
		protected.GET("/notifications", notify.List)
		protected.POST("/notifications/read", notify.MarkRead)
		protected.POST("/notifications/push-subscription", notify.SaveSubscription)
		protected.DELETE("/notifications/push-subscription", notify.RemoveSubscription)
		protected.GET("/dashboard/stats", middleware.RequireAccess(db, "dashboard", "read"), dashboard.NewHandler(db).Stats)
		accessHandler := access.NewHandler(db)
		protected.GET("/accesses", middleware.RequireAccess(db, "roles", "read"), accessHandler.Catalog)
		protected.GET("/roles/:id/accesses", middleware.RequireAccess(db, "roles", "read"), accessHandler.ListForRole)
		protected.PUT("/roles/:id/accesses", middleware.RequireAccess(db, "roles", "update"), accessHandler.ReplaceForRole)
		roles.RegisterRoutes(protected, db)
		users.RegisterRoutes(protected, db)
		settings.RegisterRoutes(protected, db)
		customers.RegisterRoutes(protected, db)
		services.RegisterRoutes(protected, db)
		invoicesessions.RegisterRoutes(protected, db)
		templates.RegisterRoutes(protected, db)
		invoices.RegisterRoutes(protected, db, emailClient)
		payments.RegisterRoutes(protected, db)
		expenses.RegisterRoutes(protected, db)
		reports.RegisterRoutes(protected, db)
		invoicepdf.RegisterRoutes(protected, db)
		whatsapp.RegisterRoutes(protected, db, whatsAppAPIURL, whatsAppAPIKey)
	}

	// SuperAdmin
	superAdmin := router.Group("/superadmin")
	superAdmin.Use(middleware.RateLimit(authRateLimiter))
	{
		superadminAuth.RegisterRoutes(superAdmin, db, superJWT, authRateLimiter, superAdminIPAllowlist, environment, bootstrapSecret)

		// Sub-groups for cleaner paths
		orgsGroup := superAdmin.Group("/organisations")
		usersGroup := superAdmin.Group("/users")

		superadminOrgs.RegisterRoutes(superAdmin, db)
		superadminUsers.RegisterRoutes(orgsGroup, usersGroup, db)
	}
}
