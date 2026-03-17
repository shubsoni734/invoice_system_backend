package templates

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	templatesdb "github.com/your-org/invoice-backend/internal/domain/templates/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := templatesdb.New(db)
	handler := NewHandler(q)

	templatesGroup := router.Group("/templates")
	{
		templatesGroup.GET("", handler.GetTemplates)
		templatesGroup.GET("/:id", handler.GetTemplateByID)
		templatesGroup.POST("", handler.CreateTemplate)
		templatesGroup.PUT("/:id", handler.UpdateTemplate)
		templatesGroup.DELETE("/:id", handler.DeleteTemplate)
	}
}
