package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

func PlanLimit(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		orgID, _ := c.Get(constants.CtxOrgID)

		var maxInvoices, maxCustomers, maxUsers int
		err := db.QueryRow(context.Background(), `
			SELECT p.max_invoices_per_month, p.max_customers, p.max_users
			FROM organisation_subscriptions os
			JOIN plans p ON os.plan_id = p.id
			WHERE os.organisation_id = $1 AND os.status = 'active'
		`, orgID).Scan(&maxInvoices, &maxCustomers, &maxUsers)
		if err != nil {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		if path == "/api/v1/invoices" {
			var count int
			_ = db.QueryRow(context.Background(), `
				SELECT COUNT(*) FROM invoices
				WHERE organisation_id = $1 AND created_at >= date_trunc('month', NOW())
			`, orgID).Scan(&count)
			if count >= maxInvoices {
				response.Error(c, http.StatusForbidden, "Monthly invoice limit reached")
				c.Abort()
				return
			}
		}

		if path == "/api/v1/customers" {
			var count int
			_ = db.QueryRow(context.Background(), `
				SELECT COUNT(*) FROM customers
				WHERE organisation_id = $1 AND is_active = TRUE
			`, orgID).Scan(&count)
			if count >= maxCustomers {
				response.Error(c, http.StatusForbidden, "Customer limit reached")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
