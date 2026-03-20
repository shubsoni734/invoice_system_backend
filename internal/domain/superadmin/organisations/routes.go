package organisations

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/pkg/email"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool, emailClient *email.Client) {
	handler := NewHandler(db, emailClient)

	router.POST("", handler.CreateOrganisation)
	router.GET("", handler.ListOrganisations)
	router.GET("/:id", handler.GetOrganisation)
	router.POST("/:id/subscription", handler.ApplySubscription)
}
