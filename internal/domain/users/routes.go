package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	orgusersdb "github.com/your-org/invoice-backend/internal/domain/users/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(protected *gin.RouterGroup, db *pgxpool.Pool) {
	q := orgusersdb.New(db)
	handler := NewHandler(q)

	users := protected.Group("/users")
	{
		users.GET("", middleware.RequireAccess(db, "users", "read"), handler.ListUsers)
		users.POST("", middleware.RequireAccess(db, "users", "create"), handler.CreateUser)
		users.PUT("/:id", middleware.RequireAccess(db, "users", "update"), handler.UpdateUser)
		users.PUT("/:id/status", middleware.RequireAccess(db, "users", "update"), handler.SetUserStatus)
	}
}
