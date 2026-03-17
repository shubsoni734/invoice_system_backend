package customers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := customersdb.New(db)
	handler := NewHandler(q)

	customersGroup := router.Group("/customers")
	{
		customersGroup.GET("", handler.GetCustomers)
		customersGroup.GET("/:id", handler.GetCustomerByID)
		customersGroup.POST("", handler.CreateCustomer)
		customersGroup.PUT("/:id", handler.UpdateCustomer)
		customersGroup.DELETE("/:id", handler.DeleteCustomer)
	}
}
