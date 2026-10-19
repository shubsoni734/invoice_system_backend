package roles

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	rolesdb "github.com/your-org/invoice-backend/internal/domain/roles/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(
	protected *gin.RouterGroup,
	db *pgxpool.Pool,
) {
	q := rolesdb.New(db)
	handler := NewHandler(q)

	roles := protected.Group("/roles")
	{
		roles.GET("", middleware.RequireAccess(db, "roles", "read"), handler.ListRoles)
		roles.GET("/:id", middleware.RequireAccess(db, "roles", "read"), handler.GetRole)
		roles.POST("", middleware.RequireAccess(db, "roles", "create"), handler.CreateRole)
		roles.PUT("/:id", middleware.RequireAccess(db, "roles", "update"), handler.UpdateRole)
		roles.DELETE("/:id", middleware.RequireAccess(db, "roles", "delete"), handler.DeleteRole)
	}
}
