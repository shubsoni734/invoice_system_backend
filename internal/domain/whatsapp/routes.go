package whatsapp

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	whatsappdb "github.com/your-org/invoice-backend/internal/domain/whatsapp/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, apiURL, apiKey string) {
	q := whatsappdb.New(db)
	handler := NewHandler(q, db, apiURL, apiKey)

	wa := router.Group("/whatsapp")
	{
		wa.GET("/logs", middleware.RequireAccess(db, "whatsapp", "read"), handler.GetWhatsAppLogs)
		wa.POST("/send", middleware.RequireAccess(db, "whatsapp", "create"), handler.SendWhatsApp)
	}
}
