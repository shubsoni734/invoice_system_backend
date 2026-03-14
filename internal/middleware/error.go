package middleware

import (
	"errors"
	"github.com/your-org/invoice-backend/internal/constants"
	"github.com/your-org/invoice-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			statusCode := mapErrorToStatusCode(err)
			utils.ErrorResponse(c, statusCode, err.Error())
		}
	}
}

func mapErrorToStatusCode(err error) int {
	switch {
	case errors.Is(err, constants.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, constants.ErrUnauthorised):
		return http.StatusUnauthorized
	case errors.Is(err, constants.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, constants.ErrValidation):
		return http.StatusUnprocessableEntity
	case errors.Is(err, constants.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, constants.ErrPlanLimit):
		return http.StatusForbidden
	case errors.Is(err, constants.ErrOrgSuspended):
		return http.StatusForbidden
	case errors.Is(err, constants.ErrAccountLocked):
		return http.StatusLocked
	case errors.Is(err, constants.ErrTokenExpired):
		return http.StatusUnauthorized
	case errors.Is(err, constants.ErrTokenInvalid):
		return http.StatusUnauthorized
	case errors.Is(err, constants.ErrMaintenance):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
