package templates

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	templatesdb "github.com/your-org/invoice-backend/internal/domain/templates/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := templatesdb.New(db)
	handler := NewHandler(q)

	templatesGroup := router.Group("/templates")
	{
		templatesGroup.GET("", middleware.RequireAccess(db, "settings", "read"), handler.GetTemplates)
		templatesGroup.GET("/:id", middleware.RequireAccess(db, "settings", "read"), handler.GetTemplateByID)
		templatesGroup.POST("", middleware.RequireAccess(db, "settings", "update"), handler.CreateTemplate)
		templatesGroup.PUT("/:id", middleware.RequireAccess(db, "settings", "update"), handler.UpdateTemplate)
		templatesGroup.DELETE("/:id", middleware.RequireAccess(db, "settings", "update"), handler.DeleteTemplate)
	}
}
