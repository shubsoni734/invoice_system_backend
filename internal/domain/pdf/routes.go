package pdf

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	handler := NewHandler(db)
	router.GET("/invoices/:id/pdf", handler.GenerateInvoicePDF)
}
