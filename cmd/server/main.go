package main

import (
	"context"
	"fmt"
	"github.com/your-org/invoice-backend/internal/config"
	"github.com/your-org/invoice-backend/internal/middleware"
	"github.com/your-org/invoice-backend/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	superadminAuth "github.com/your-org/invoice-backend/internal/modules/superadmin/auth"
	superadminOrgs "github.com/your-org/invoice-backend/internal/modules/superadmin/organisations"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	var logger *zap.Logger
	if cfg.Logging.Format == "json" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	dbConfig, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to parse database config", zap.Error(err))
	}

	dbConfig.MinConns = int32(cfg.Database.MinConns)
	dbConfig.MaxConns = int32(cfg.Database.MaxConns)
	dbConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	dbConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime

	dbPool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Database connection established")

	orgJWTManager, err := utils.NewJWTManager(
		cfg.OrgJWT.PrivateKeyPath,
		cfg.OrgJWT.PublicKeyPath,
		cfg.OrgJWT.AccessTokenExpiry,
	)
	if err != nil {
		logger.Fatal("Failed to initialize org JWT manager", zap.Error(err))
	}

	superJWTManager, err := utils.NewJWTManager(
		cfg.SuperJWT.PrivateKeyPath,
		cfg.SuperJWT.PublicKeyPath,
		cfg.SuperJWT.AccessTokenExpiry,
	)
	if err != nil {
		logger.Fatal("Failed to initialize super JWT manager", zap.Error(err))
	}

	authRateLimiter := middleware.NewRateLimiter(cfg.RateLimit.AuthRPM)
	apiRateLimiter := middleware.NewRateLimiter(cfg.RateLimit.APIRPM)

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, "Server is running", nil)
	})

	router.GET("/ready", func(c *gin.Context) {
		if err := dbPool.Ping(c.Request.Context()); err != nil {
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "Database not ready")
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Server is ready", nil)
	})

	apiV1 := router.Group("/api/v1")
	{
		authPublic := apiV1.Group("/auth")
		authPublic.Use(middleware.RateLimit(authRateLimiter))
		{
			authPublic.POST("/login", func(c *gin.Context) {
				utils.SuccessResponse(c, http.StatusOK, "Login endpoint - implement auth module", nil)
			})
		}

		protected := apiV1.Group("")
		protected.Use(middleware.RateLimit(apiRateLimiter))
		protected.Use(middleware.Auth(orgJWTManager))
		protected.Use(middleware.Tenant(dbPool))
		{
			protected.GET("/auth/me", func(c *gin.Context) {
				utils.SuccessResponse(c, http.StatusOK, "User profile endpoint", nil)
			})
		}
	}

	superAdmin := router.Group("/superadmin")
	superAdmin.Use(middleware.RateLimit(authRateLimiter))
	{
		// Public: create superadmin + login
		superadminAuth.RegisterRoutes(superAdmin, dbPool, superJWTManager, authRateLimiter)

		// Protected: requires valid superadmin JWT
		protected := superAdmin.Group("")
		protected.Use(middleware.SuperAuth(superJWTManager, cfg.SuperAdmin.IPAllowlist))
		{
			superadminOrgs.RegisterRoutes(protected.Group("/organisations"), dbPool)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	dbPool.Close()
	logger.Info("Server exited cleanly")
}
