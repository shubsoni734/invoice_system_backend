package customers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := customersdb.New(db)
	handler := NewHandler(q)

	customersGroup := router.Group("/customers")
	{
		customersGroup.GET("", middleware.RequireAccess(db, "customers", "read"), handler.GetCustomers)
		customersGroup.GET("/:id", middleware.RequireAccess(db, "customers", "read"), handler.GetCustomerByID)
		customersGroup.POST("", middleware.RequireAccess(db, "customers", "create"), handler.CreateCustomer)
		customersGroup.PUT("/:id", middleware.RequireAccess(db, "customers", "update"), handler.UpdateCustomer)
		customersGroup.DELETE("/:id", middleware.RequireAccess(db, "customers", "delete"), handler.DeleteCustomer)
	}
}
