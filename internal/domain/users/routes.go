package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	orgusersdb "github.com/your-org/invoice-backend/internal/domain/users/sqlc"
)

func RegisterRoutes(protected *gin.RouterGroup, db *pgxpool.Pool) {
	q := orgusersdb.New(db)
	handler := NewHandler(q)

	users := protected.Group("/users")
	{
		users.GET("", handler.ListUsers)
		users.POST("", handler.CreateUser)
		users.PUT("/:id/status", handler.SetUserStatus)
	}
}
