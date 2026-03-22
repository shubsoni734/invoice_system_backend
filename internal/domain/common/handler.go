package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/pkg/email"
	"github.com/your-org/invoice-backend/internal/pkg/response"
)

type Handler struct {
	emailClient *email.Client
}

func NewHandler(emailClient *email.Client) *Handler {
	return &Handler{emailClient: emailClient}
}

type TestEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// TestEmail POST /api/v1/common/test-email
func (h *Handler) TestEmail(c *gin.Context) {
	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// For test email, we can send a simple welcome-like message
	err := h.emailClient.SendWelcomeEmail(req.Email, "Test User", "TEST_PASSWORD_123")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send test email: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Test email sent successfully to "+req.Email, nil)
}
