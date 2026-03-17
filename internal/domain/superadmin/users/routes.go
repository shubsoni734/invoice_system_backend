package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	usersdb "github.com/your-org/invoice-backend/internal/domain/superadmin/users/sqlc"
)

func RegisterRoutes(orgsGroup *gin.RouterGroup, usersGroup *gin.RouterGroup, db *pgxpool.Pool) {
	q := usersdb.New(db)
	handler := NewHandler(q)

	// Under /superadmin/organisations/:id/users
	orgsGroup.GET("/:id/users", handler.ListUsers)
	orgsGroup.POST("/:id/users", handler.CreateUser)

	// Under /superadmin/users/:id/status
	usersGroup.PUT("/:id/status", handler.SetUserStatus)
}
