package services

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	servicesdb "github.com/your-org/invoice-backend/internal/domain/services/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := servicesdb.New(db)
	handler := NewHandler(q)

	servicesGroup := router.Group("/services")
	{
		servicesGroup.GET("", handler.GetServices)
		servicesGroup.GET("/:id", handler.GetServiceByID)
		servicesGroup.POST("", handler.CreateService)
		servicesGroup.PUT("/:id", handler.UpdateService)
		servicesGroup.DELETE("/:id", handler.DeleteService)
	}
}
