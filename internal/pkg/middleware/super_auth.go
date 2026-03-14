package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

func SuperAuth(jwtManager *utils.JWTManager, ipAllowlist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(ipAllowlist) > 0 {
			clientIP := c.ClientIP()
			allowed := false
			for _, ip := range ipAllowlist {
				if clientIP == ip {
					allowed = true
					break
				}
			}
			if !allowed {
				response.Error(c, http.StatusForbidden, "IP not allowed")
				c.Abort()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.VerifyToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		if claims.SuperAdminID == "" {
			response.Error(c, http.StatusUnauthorized, "Not a superadmin token")
			c.Abort()
			return
		}

		c.Set(constants.CtxSuperAdminID, claims.SuperAdminID)
		c.Set(constants.CtxSuperAdminRole, claims.Role)
		c.Next()
	}
}
