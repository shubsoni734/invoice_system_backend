package settings

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := settingsdb.New(db)
	handler := NewHandler(q, db)

	router.GET("/settings", middleware.RequireAccess(db, "settings", "read"), handler.GetSettings)
	router.PUT("/settings", middleware.RequireAccess(db, "settings", "update"), handler.UpsertSettings)
	router.GET("/settings/channels", middleware.RequireAccess(db, "settings", "read"), handler.GetChannels)
	router.PUT("/settings/channels", middleware.RequireAccess(db, "settings", "update"), handler.UpdateChannels)
	router.POST("/settings/channels/test", middleware.RequireAccess(db, "settings", "update"), handler.TestChannelEmail)
}
