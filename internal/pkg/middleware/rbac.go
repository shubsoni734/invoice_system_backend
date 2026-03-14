package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

func RBAC(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(constants.CtxUserRole)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User role not found")
			c.Abort()
			return
		}
		role := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, "Insufficient permissions")
		c.Abort()
	}
}

func SuperRBAC(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		superAdminRole, exists := c.Get(constants.CtxSuperAdminRole)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "SuperAdmin role not found")
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
		response.Error(c, http.StatusForbidden, "Insufficient permissions")
		c.Abort()
	}
}
