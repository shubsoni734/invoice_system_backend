package organisations

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	orgdb "github.com/your-org/invoice-backend/internal/domain/superadmin/organisations/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := orgdb.New(db)
	handler := NewHandler(q)

	router.POST("/organisations", handler.CreateOrganisation)
	router.GET("/organisations", handler.ListOrganisations)
	router.GET("/organisations/:id", handler.GetOrganisationByID)
	router.POST("/organisations/:id/subscription", handler.ApplySubscription)
}
