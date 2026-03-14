package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/constants"
	"github.com/your-org/invoice-backend/internal/utils"
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
				utils.ErrorResponse(c, http.StatusForbidden, "IP not allowed")
				c.Abort()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.VerifyToken(parts[1])
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		if claims.SuperAdminID == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Not a superadmin token")
			c.Abort()
			return
		}

		c.Set(constants.CtxSuperAdminID, claims.SuperAdminID)
		c.Set(constants.CtxSuperAdminRole, claims.Role)
		c.Next()
	}
}
