package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/domain/access"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

// RequireAccess allows the request only when the signed-in user's role
// includes the page action.
// Check is one page action that can satisfy a route.
type Check struct {
	Resource string
	Action   string
}

func RequireAccess(db *pgxpool.Pool, resource, action string) gin.HandlerFunc {
	return requireAny(db, Check{Resource: resource, Action: action})
}

// RequireAnyAccess allows the request when any one of the checks passes.
func RequireAnyAccess(db *pgxpool.Pool, checks ...Check) gin.HandlerFunc {
	return requireAny(db, checks...)
}

func requireAny(db *pgxpool.Pool, checks ...Check) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.GetString(constants.CtxUserID))
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		for _, check := range checks {
			allowed, err := access.Allowed(c.Request.Context(), db, userID, check.Resource, check.Action)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check access")
				c.Abort()
				return
			}
			if allowed {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, "You do not have access to this action")
		c.Abort()
	}
}
