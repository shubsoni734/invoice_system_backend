package organisations

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	handler := NewHandler(db)

	router.POST("", handler.CreateOrganisation)
	router.GET("", handler.ListOrganisations)
	router.GET("/:id", handler.GetOrganisation)
	router.POST("/:id/subscription", handler.ApplySubscription)
}
