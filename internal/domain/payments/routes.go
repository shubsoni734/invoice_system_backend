package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	paymentsdb "github.com/your-org/invoice-backend/internal/domain/payments/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := paymentsdb.New(db)
	handler := NewHandler(q, db)

	router.GET("/payments", middleware.RequireAccess(db, "payments", "read"), handler.ListPayments)
	router.POST("/payments", middleware.RequireAccess(db, "payments", "create"), handler.RecordPayment)
	router.PUT("/payments/:id", middleware.RequireAccess(db, "payments", "update"), handler.UpdatePayment)
	router.DELETE("/payments/:id", middleware.RequireAccess(db, "payments", "delete"), handler.DeletePayment)
	router.GET("/invoices/:id/payments", middleware.RequireAccess(db, "invoices", "read"), handler.GetPaymentsByInvoice)
}
