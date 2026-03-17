package whatsapp

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	whatsappdb "github.com/your-org/invoice-backend/internal/domain/whatsapp/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, apiURL, apiKey string) {
	q := whatsappdb.New(db)
	handler := NewHandler(q, db, apiURL, apiKey)

	wa := router.Group("/whatsapp")
	{
		wa.GET("/logs", handler.GetWhatsAppLogs)
		wa.POST("/send", handler.SendWhatsApp)
	}
}
