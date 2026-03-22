package roles

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	rolesdb "github.com/your-org/invoice-backend/internal/domain/roles/sqlc"
)

func RegisterRoutes(
	protected *gin.RouterGroup,
	db *pgxpool.Pool,
) {
	q := rolesdb.New(db)
	handler := NewHandler(q)

	roles := protected.Group("/roles")
	{
		roles.GET("", handler.ListRoles)
		roles.GET("/:id", handler.GetRole)
		roles.POST("", handler.CreateRole)
		roles.PUT("/:id", handler.UpdateRole)
		roles.DELETE("/:id", handler.DeleteRole)
	}
}
