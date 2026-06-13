package invoices

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/email"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, emailClient *email.Client) {
	q := invoicesdb.New(db)
	handler := NewHandler(q, db, emailClient)

	invoicesGroup := router.Group("/invoices")
	{
		invoicesGroup.GET("", handler.GetInvoices)
		invoicesGroup.GET("/:id", handler.GetInvoiceByID)
		invoicesGroup.POST("", handler.CreateInvoice)
		invoicesGroup.PUT("/:id", handler.UpdateInvoice)
		invoicesGroup.PUT("/:id/send", handler.SendInvoice)
		invoicesGroup.PUT("/:id/paid", handler.MarkInvoiceAsPaid)
		invoicesGroup.PUT("/:id/cancel", handler.CancelInvoice)
		invoicesGroup.DELETE("/:id", handler.DeleteInvoice)
		invoicesGroup.POST("/:id/email", handler.SendInvoiceEmail)
	}
}
