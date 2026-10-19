package invoicesessions

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	invoicesessionsdb "github.com/your-org/invoice-backend/internal/domain/invoicesessions/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := invoicesessionsdb.New(db)
	handler := NewHandler(q)

	sessionGroup := router.Group("/invoice-sessions")
	{
		sessionGroup.GET("", middleware.RequireAccess(db, "invoices", "read"), handler.GetInvoiceSessions)
		sessionGroup.GET("/:id", middleware.RequireAccess(db, "invoices", "read"), handler.GetInvoiceSessionByID)
		sessionGroup.POST("", middleware.RequireAccess(db, "invoices", "create"), handler.CreateInvoiceSession)
	}
}
