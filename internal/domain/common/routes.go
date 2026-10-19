package common

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/pkg/email"
)

func RegisterRoutes(
	public *gin.RouterGroup,
	emailClient *email.Client,
) {
	handler := NewHandler(emailClient)

	common := public.Group("/common")
	{
		common.POST("/test-email", handler.TestEmail)
	}
}

// RegisterSampleRoute exposes the secret email-check page API.
// An empty token leaves the route unregistered.
func RegisterSampleRoute(public *gin.RouterGroup, emailClient *email.Client, token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	handler := NewHandler(emailClient)
	handler.testToken = token
	public.POST("/email-check/:token", handler.SendSample)
}
