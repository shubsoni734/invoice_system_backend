package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	paymentsdb "github.com/your-org/invoice-backend/internal/domain/payments/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := paymentsdb.New(db)
	handler := NewHandler(q)

	// Record a payment
	router.POST("/payments", handler.RecordPayment)
	// Delete a payment
	router.DELETE("/payments/:id", handler.DeletePayment)
	// Get payments for a specific invoice
	router.GET("/invoices/:id/payments", handler.GetPaymentsByInvoice)
}
