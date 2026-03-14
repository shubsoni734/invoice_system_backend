package middleware

import (
	"github.com/your-org/invoice-backend/internal/constants"
	"github.com/your-org/invoice-backend/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		token := parts[1]
		claims, err := jwtManager.VerifyToken(token)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set(constants.CtxUserID, claims.UserID)
		c.Set(constants.CtxOrgID, claims.OrgID)
		c.Set(constants.CtxUserRole, claims.Role)

		if claims.ImpersonatedBy > 0 {
			c.Set(constants.CtxIsImpersonating, true)
		}

		c.Next()
	}
}
