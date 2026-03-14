package middleware

import (
	"github.com/your-org/invoice-backend/internal/constants"
	"github.com/your-org/invoice-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SuperRBAC(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		superAdminRole, exists := c.Get(constants.CtxSuperAdminRole)
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "SuperAdmin role not found")
			c.Abort()
			return
		}

		role := superAdminRole.(string)
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions")
		c.Abort()
	}
}
