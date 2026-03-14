package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

func Tenant(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, exists := c.Get(constants.CtxOrgID)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Organisation ID not found")
			c.Abort()
			return
		}

		var status string
		err := db.QueryRow(context.Background(),
			"SELECT status FROM organisations WHERE id = $1", orgID,
		).Scan(&status)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to verify organisation")
			c.Abort()
			return
		}

		if status != constants.OrgActive {
			response.Error(c, http.StatusForbidden, "Organisation is not active")
			c.Abort()
			return
		}

		c.Next()
	}
}
