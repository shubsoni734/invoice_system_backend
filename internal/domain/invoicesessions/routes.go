package invoicesessions

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	invoicesessionsdb "github.com/your-org/invoice-backend/internal/domain/invoicesessions/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := invoicesessionsdb.New(db)
	handler := NewHandler(q)

	sessionGroup := router.Group("/invoice-sessions")
	{
		sessionGroup.GET("", handler.GetInvoiceSessions)
		sessionGroup.GET("/:id", handler.GetInvoiceSessionByID)
		sessionGroup.POST("", handler.CreateInvoiceSession)
	}
}
