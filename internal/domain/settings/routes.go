package settings

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := settingsdb.New(db)
	handler := NewHandler(q)

	router.GET("/settings", handler.GetSettings)
	router.PUT("/settings", handler.UpsertSettings)
}
