package services

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	servicesdb "github.com/your-org/invoice-backend/internal/domain/services/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := servicesdb.New(db)
	handler := NewHandler(q)

	servicesGroup := router.Group("/services")
	{
		servicesGroup.GET("", middleware.RequireAccess(db, "services", "read"), handler.GetServices)
		servicesGroup.GET("/:id", middleware.RequireAccess(db, "services", "read"), handler.GetServiceByID)
		servicesGroup.POST("", middleware.RequireAccess(db, "services", "create"), handler.CreateService)
		servicesGroup.PUT("/:id", middleware.RequireAccess(db, "services", "update"), handler.UpdateService)
		servicesGroup.DELETE("/:id", middleware.RequireAccess(db, "services", "delete"), handler.DeleteService)
	}
}
