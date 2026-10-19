package invoices

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/email"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, emailClient *email.Client) {
	q := invoicesdb.New(db)
	handler := NewHandler(q, db, emailClient)

	invoicesGroup := router.Group("/invoices")
	{
		invoicesGroup.GET("", middleware.RequireAccess(db, "invoices", "read"), handler.GetInvoices)
		invoicesGroup.GET("/:id", middleware.RequireAccess(db, "invoices", "read"), handler.GetInvoiceByID)
		invoicesGroup.POST("", middleware.RequireAccess(db, "invoices", "create"), handler.CreateInvoice)
		invoicesGroup.PUT("/:id", middleware.RequireAccess(db, "invoices", "update"), handler.UpdateInvoice)
		invoicesGroup.PUT("/:id/send", middleware.RequireAccess(db, "invoices", "update"), handler.SendInvoice)
		invoicesGroup.PUT("/:id/paid", middleware.RequireAccess(db, "invoices", "update"), handler.MarkInvoiceAsPaid)
		invoicesGroup.PUT("/:id/cancel", middleware.RequireAccess(db, "invoices", "update"), handler.CancelInvoice)
		invoicesGroup.DELETE("/:id", middleware.RequireAccess(db, "invoices", "delete"), handler.DeleteInvoice)
		invoicesGroup.POST("/:id/email", middleware.RequireAccess(db, "invoices", "update"), handler.SendInvoiceEmail)
	}
}
