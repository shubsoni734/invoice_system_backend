package expenses

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	handler := NewHandler(db)
	group := router.Group("/expenses")
	group.GET("", middleware.RequireAccess(db, "expenses", "read"), handler.List)
	group.POST("", middleware.RequireAccess(db, "expenses", "create"), handler.Create)
	group.PUT("/:id", middleware.RequireAccess(db, "expenses", "update"), handler.Update)
	group.DELETE("/:id", middleware.RequireAccess(db, "expenses", "delete"), handler.Delete)
}
