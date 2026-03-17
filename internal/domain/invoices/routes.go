package invoices

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := invoicesdb.New(db)
	handler := NewHandler(q)

	invoicesGroup := router.Group("/invoices")
	{
		invoicesGroup.GET("", handler.GetInvoices)
		invoicesGroup.GET("/:id", handler.GetInvoiceByID)
		invoicesGroup.POST("", handler.CreateInvoice)
		invoicesGroup.PUT("/:id/cancel", handler.CancelInvoice)
	}
}
