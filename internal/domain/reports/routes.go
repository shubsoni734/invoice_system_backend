package reports

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	reportsdb "github.com/your-org/invoice-backend/internal/domain/reports/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := reportsdb.New(db)
	handler := NewHandler(q)

	reportsGroup := router.Group("/reports")
	{
		reportsGroup.GET("/daily", handler.GetDailyReport)
		reportsGroup.GET("/monthly", handler.GetMonthlyReport)
		reportsGroup.GET("/customer/:id", handler.GetCustomerReport)
		reportsGroup.GET("/revenue", handler.GetRevenueSummary)
	}
}
