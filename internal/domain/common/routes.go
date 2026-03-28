package common

import (
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
