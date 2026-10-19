package pdf

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	handler := NewHandler(db)
	router.GET("/invoices/:id/pdf", middleware.RequireAccess(db, "invoices", "read"), handler.GenerateInvoicePDF)
}
