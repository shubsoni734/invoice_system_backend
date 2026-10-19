package reports

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	reportsdb "github.com/your-org/invoice-backend/internal/domain/reports/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *pgxpool.Pool) {
	q := reportsdb.New(db)
	handler := NewHandler(q, db)

	readReport := middleware.RequireAnyAccess(db,
		middleware.Check{Resource: "reports", Action: "read"},
		middleware.Check{Resource: "gst", Action: "read"},
	)
	reportsGroup := router.Group("/reports")
	{
		reportsGroup.GET("/daily", readReport, handler.GetDailyReport)
		reportsGroup.GET("/monthly", readReport, handler.GetMonthlyReport)
		reportsGroup.GET("/customer/:id", readReport, handler.GetCustomerReport)
		reportsGroup.GET("/revenue", readReport, handler.GetRevenueSummary)
		reportsGroup.GET("/overview", readReport, handler.GetOverview)
		reportsGroup.GET("/gst", readReport, handler.GetGST)
	}
}
